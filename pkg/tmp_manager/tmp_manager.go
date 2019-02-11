package tmp_manager

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/flant/werf/pkg/dappdeps"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/logger"
	"github.com/flant/werf/pkg/werf"
	"github.com/otiai10/copy"
)

const (
	projectsDir      = "projects"
	dockerConfigsDir = "docker_configs"
)

func GetCreatedTmpDirs() string {
	return filepath.Join(werf.GetServiceDir(), "tmp", "created")
}

func GetReleasedTmpDirs() string {
	return filepath.Join(werf.GetServiceDir(), "tmp", "released")
}

func registerCreatedDir(newDir, createdDirs string) error {
	if err := os.MkdirAll(createdDirs, os.ModePerm); err != nil {
		return fmt.Errorf("unable to create dir %s: %s", createdDirs, err)
	}

	createdDir := filepath.Join(createdDirs, filepath.Base(newDir))
	if err := os.Symlink(newDir, createdDir); err != nil {
		return err
	}

	return nil
}

func releaseDir(dir, createdDirs, releasedDirs string) error {
	if err := os.MkdirAll(releasedDirs, os.ModePerm); err != nil {
		return fmt.Errorf("unable to create dir %s: %s", releaseDir, err)
	}

	releasedDir := filepath.Join(releasedDirs, filepath.Base(dir))
	if err := os.Symlink(dir, releasedDir); err != nil {
		return err
	}

	createdDir := filepath.Join(createdDirs, filepath.Base(dir))
	if err := os.Remove(createdDir); err != nil {
		return err
	}

	return nil
}

func checkShouldRunGC() (bool, error) {
	releasedProjectsDir := filepath.Join(GetReleasedTmpDirs(), projectsDir)
	if _, err := os.Stat(releasedProjectsDir); !os.IsNotExist(err) {
		var err error
		releasedProjectsDirs, err := ioutil.ReadDir(releasedProjectsDir)
		if err != nil {
			return false, fmt.Errorf("unable to list released projects tmp dirs in %s: %s", releasedProjectsDir, err)
		}

		if len(releasedProjectsDirs) > 50 {
			return true, nil
		}
	}

	createdDockerConfigsDir := filepath.Join(GetCreatedTmpDirs(), dockerConfigsDir)
	if _, err := os.Stat(createdDockerConfigsDir); !os.IsNotExist(err) {
		var err error
		createdDockerConfigsDirs, err := ioutil.ReadDir(createdDockerConfigsDir)
		if err != nil {
			return false, fmt.Errorf("unable to list released docker configs tmp dirs in %s: %s", createdDockerConfigsDir, err)
		}

		now := time.Now()
		for _, info := range createdDockerConfigsDirs {
			if now.Sub(info.ModTime()) > 24*time.Hour {
				return true, nil
			}
		}
	}

	return false, nil
}

func newTmpDir(prefix string) (string, error) {
	newDir, err := ioutil.TempDir(werf.GetTmpDir(), prefix)
	if err != nil {
		return "", err
	}

	if runtime.GOOS == "darwin" {
		dir, err := filepath.EvalSymlinks(newDir)
		if err != nil {
			return "", fmt.Errorf("eval symlink %s failed: %s", newDir, err)
		}
		newDir = dir
	}

	return newDir, nil
}

func CreateProjectDir() (string, error) {
	newDir, err := newTmpDir("werf-project-data-")
	if err != nil {
		return "", err
	}

	if err := registerCreatedDir(newDir, filepath.Join(GetCreatedTmpDirs(), projectsDir)); err != nil {
		os.RemoveAll(newDir)
		return "", err
	}

	shouldRunGC, err := checkShouldRunGC()
	if err != nil {
		return "", err
	}

	if shouldRunGC {
		err := lock.WithLock("gc", lock.LockOptions{}, GC)
		if err != nil {
			return "", fmt.Errorf("GC failed: %s", err)
		}
	}

	return newDir, nil
}

func ReleaseProjectDir(dir string) error {
	return releaseDir(dir, filepath.Join(GetCreatedTmpDirs(), projectsDir), filepath.Join(GetReleasedTmpDirs(), projectsDir))
}

func CreateDockerConfigDir(fromDockerConfig string) (string, error) {
	newDir, err := newTmpDir("werf-docker-config-")
	if err != nil {
		return "", err
	}

	if err := os.Chmod(newDir, 0700); err != nil {
		return "", err
	}

	if _, err := os.Stat(fromDockerConfig); !os.IsNotExist(err) {
		err := copy.Copy(fromDockerConfig, newDir)
		if err != nil {
			return "", fmt.Errorf("unable to copy %s to %s: %s", fromDockerConfig, newDir, err)
		}
	}

	if err := registerCreatedDir(newDir, filepath.Join(GetCreatedTmpDirs(), dockerConfigsDir)); err != nil {
		os.RemoveAll(newDir)
		return "", err
	}

	shouldRunGC, err := checkShouldRunGC()
	if err != nil {
		return "", err
	}

	if shouldRunGC {
		err := lock.WithLock("gc", lock.LockOptions{}, GC)
		if err != nil {
			return "", fmt.Errorf("GC failed: %s", err)
		}
	}

	return newDir, nil
}

type DirDesc struct {
	FileInfo os.FileInfo
	LinkPath string
}

func getDirsDescs(dir string) ([]*DirDesc, error) {
	var res []*DirDesc

	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		infos, err := ioutil.ReadDir(dir)
		if err != nil {
			return nil, fmt.Errorf("unable to list dirs in %s: %s", dir, err)
		}

		for _, info := range infos {
			res = append(res, &DirDesc{FileInfo: info, LinkPath: filepath.Join(dir, info.Name())})
		}
	}

	return res, nil
}

func getReleasedDirsToRemove(releasedDirs []*DirDesc) ([]string, error) {
	var res []string

	for _, desc := range releasedDirs {
		origDir, err := os.Readlink(desc.LinkPath)
		if err != nil {
			return nil, fmt.Errorf("unable to read link %s: %s", desc.LinkPath, err)
		}
		res = append(res, origDir, desc.LinkPath)
	}

	return res, nil
}

func getCreatedDirsToRemove(createdDirs []*DirDesc) ([]string, error) {
	var res []string

	now := time.Now()
	for _, desc := range createdDirs {
		if now.Sub(desc.FileInfo.ModTime()) < 2*time.Hour {
			continue
		}

		origDir, err := os.Readlink(desc.LinkPath)
		if err != nil {
			return nil, fmt.Errorf("unable to read link %s: %s", desc.LinkPath, err)
		}

		res = append(res, origDir, desc.LinkPath)
	}

	return res, nil
}

func GC() error {
	logger.LogInfoF("Running GC\n")

	releasedProjectsDescs, err := getDirsDescs(filepath.Join(GetReleasedTmpDirs(), projectsDir))
	if err != nil {
		return fmt.Errorf("unable to get released tmp projects dirs: %s", err)
	}

	createdProjectsDescs, err := getDirsDescs(filepath.Join(GetCreatedTmpDirs(), projectsDir))
	if err != nil {
		return fmt.Errorf("unable to get created tmp projects dirs: %s", err)
	}

	createdDockerConfigsDescs, err := getDirsDescs(filepath.Join(GetCreatedTmpDirs(), dockerConfigsDir))
	if err != nil {
		return fmt.Errorf("unable to get created tmp docker configs dirs: %s", err)
	}

	var dirs []string

	projectsToRemove := []string{}
	dockerConfigsToRemove := []string{}

	dirs, err = getReleasedDirsToRemove(releasedProjectsDescs)
	if err != nil {
		return fmt.Errorf("unable to get released tmp projects dirs to remove: %s", err)
	}
	projectsToRemove = append(projectsToRemove, dirs...)

	dirs, err = getCreatedDirsToRemove(createdProjectsDescs)
	if err != nil {
		return fmt.Errorf("unable to get created tmp projects dirs to remove: %s", err)
	}
	projectsToRemove = append(projectsToRemove, dirs...)

	dirs, err = getCreatedDirsToRemove(createdDockerConfigsDescs)
	if err != nil {
		return fmt.Errorf("unable to get created tmp docker configs dirs to remove: %s", err)
	}
	dockerConfigsToRemove = append(dockerConfigsToRemove, dirs...)

	var removeErrors []error

	if len(projectsToRemove) > 0 {
		if err := removeProjectDirs(projectsToRemove); err != nil {
			removeErrors = append(removeErrors, fmt.Errorf("unable to remove tmp projects dirs %s: %s", strings.Join(projectsToRemove, ", "), err))
		}
	}

	for _, dir := range dockerConfigsToRemove {
		err := os.RemoveAll(dir)
		if err != nil {
			removeErrors = append(removeErrors, fmt.Errorf("unable to remove %s: %s", dir, err))
		}
	}

	if len(removeErrors) > 0 {
		msg := ""
		for _, err := range removeErrors {
			msg += fmt.Sprintf("%s\n", err)
		}
		return fmt.Errorf("%s", msg)
	}

	return nil
}

func Purge() error {
	tmpFiles, err := ioutil.ReadDir(werf.GetTmpDir())
	if err != nil {
		return fmt.Errorf("unable to list tmp files in %s: %s", werf.GetTmpDir(), err)
	}

	filesToRemove := []string{}
	for _, finfo := range tmpFiles {
		if strings.HasPrefix(finfo.Name(), "werf") {
			filesToRemove = append(filesToRemove, filepath.Join(werf.GetTmpDir(), finfo.Name()))
		}
	}

	for _, file := range filesToRemove {
		err := os.RemoveAll(file)
		if err != nil {
			logger.LogErrorF("WARNING: unable to remove %s: %s\n", file, err)
		}
	}

	return nil
}

func removeProjectDirs(dirs []string) error {
	toolchainContainerName, err := dappdeps.ToolchainContainer()
	if err != nil {
		return err
	}

	args := []string{
		"--rm",
		"--volumes-from", toolchainContainerName,
		"--volume", fmt.Sprintf("%s:%s", werf.GetTmpDir(), werf.GetTmpDir()),
		dappdeps.BaseImageName(),
		dappdeps.RmBinPath(), "-rf",
	}

	args = append(args, dirs...)

	return docker.CliRun(args...)
}
