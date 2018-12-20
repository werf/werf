package project_tmp_dir

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flant/dapp/pkg/dapp"
	"github.com/flant/dapp/pkg/dappdeps"
	"github.com/flant/dapp/pkg/docker"
	"github.com/flant/dapp/pkg/lock"
)

func GetCreatedTmpDirs() string {
	return filepath.Join(dapp.GetHomeDir(), "tmp", "created")
}

func GetReleasedTmpDirs() string {
	return filepath.Join(dapp.GetHomeDir(), "tmp", "released")
}

func Get() (string, error) {
	dir, err := ioutil.TempDir(dapp.GetTmpDir(), "dapp-")
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(GetCreatedTmpDirs(), os.ModePerm); err != nil {
		return "", fmt.Errorf("unable to create dir %s: %s", GetCreatedTmpDirs(), err)
	}

	createdDir := filepath.Join(GetCreatedTmpDirs(), filepath.Base(dir))
	if err := os.Symlink(dir, createdDir); err != nil {
		os.RemoveAll(dir)
		return "", err
	}

	gcShouldRun := false

	if _, err := os.Stat(GetReleasedTmpDirs()); !os.IsNotExist(err) {
		releasedDirs, err := ioutil.ReadDir(GetReleasedTmpDirs())
		if err != nil {
			return "", fmt.Errorf("unable to list released tmp dirs in %s: %s", GetReleasedTmpDirs(), err)
		}

		if len(releasedDirs) > 50 {
			gcShouldRun = true
		}
	}

	if gcShouldRun {
		err := lock.WithLock("gc", lock.LockOptions{}, GC)
		if err != nil {
			return "", fmt.Errorf("GC failed: %s", err)
		}
	}

	return dir, nil
}

func GC() error {
	var releasedDirs []os.FileInfo
	var createdDirs []os.FileInfo

	if _, err := os.Stat(GetReleasedTmpDirs()); !os.IsNotExist(err) {
		var err error
		releasedDirs, err = ioutil.ReadDir(GetReleasedTmpDirs())
		if err != nil {
			return fmt.Errorf("unable to list released tmp dirs in %s: %s", GetReleasedTmpDirs(), err)
		}
	}

	if _, err := os.Stat(GetCreatedTmpDirs()); !os.IsNotExist(err) {
		var err error
		createdDirs, err = ioutil.ReadDir(GetCreatedTmpDirs())
		if err != nil {
			return fmt.Errorf("unable to list created tmp dirs in %s: %s", GetReleasedTmpDirs(), err)
		}
	}

	tmpDirsToRemove := []string{}
	filesToUnlink := []string{}

	for _, dirInfo := range releasedDirs {
		link := filepath.Join(GetReleasedTmpDirs(), dirInfo.Name())

		origDir, err := os.Readlink(link)
		if err != nil {
			return fmt.Errorf("unable to read link %s: %s", link, err)
		}

		tmpDirsToRemove = append(tmpDirsToRemove, origDir)
		filesToUnlink = append(filesToUnlink, link)
	}

	now := time.Now()
	for _, dirInfo := range createdDirs {
		if now.Sub(dirInfo.ModTime()) < 2*time.Hour {
			continue
		}

		link := filepath.Join(GetCreatedTmpDirs(), dirInfo.Name())

		origDir, err := os.Readlink(link)
		if err != nil {
			return fmt.Errorf("unable to read link %s: %s", link, err)
		}

		tmpDirsToRemove = append(tmpDirsToRemove, origDir)
		filesToUnlink = append(filesToUnlink, link)
	}

	if len(tmpDirsToRemove) > 0 {
		if err := removeDirs(tmpDirsToRemove); err != nil {
			return fmt.Errorf("unable to remove dirs %s: %s", strings.Join(tmpDirsToRemove, ", "), err)
		}
	}

	for _, file := range filesToUnlink {
		if err := os.Remove(file); err != nil {
			return fmt.Errorf("unable to remove %s: %s", file, err)
		}
	}

	return nil
}

func Release(dir string) error {
	if err := os.MkdirAll(GetReleasedTmpDirs(), os.ModePerm); err != nil {
		return fmt.Errorf("unable to create dir %s: %s", GetReleasedTmpDirs(), err)
	}

	releasedDir := filepath.Join(GetReleasedTmpDirs(), filepath.Base(dir))
	if err := os.Symlink(dir, releasedDir); err != nil {
		return err
	}

	createdDir := filepath.Join(GetCreatedTmpDirs(), filepath.Base(dir))
	if err := os.Remove(createdDir); err != nil {
		return err
	}

	return nil
}

func removeDirs(dirs []string) error {
	toolchainContainerName, err := dappdeps.ToolchainContainer()
	if err != nil {
		return err
	}

	args := []string{
		"--rm",
		"--volumes-from", toolchainContainerName,
		"--volume", fmt.Sprintf("%s:%s", dapp.GetTmpDir(), dapp.GetTmpDir()),
		dappdeps.BaseImageName(),
		dappdeps.RmBinPath(), "-rf",
	}

	args = append(args, dirs...)

	return docker.CliRun(args...)
}
