package tmp_manager

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/werf/lockgate"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
)

var (
	AutoGCEnabled = os.Getenv("WERF_DISABLE_AUTO_GC") != "1"
)

func runGC() error {
	return werf.WithHostLock("gc", lockgate.AcquireOptions{}, func() error {
		return GC(false)
	})
}

func checkShouldRunGC() (bool, error) {
	if !AutoGCEnabled {
		return false, nil
	}

	releasedProjectsDir := filepath.Join(GetReleasedTmpDirs(), projectsServiceDir)
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

	now := time.Now()

	createdDockerConfigsDir := filepath.Join(GetCreatedTmpDirs(), dockerConfigsServiceDir)
	if _, err := os.Stat(createdDockerConfigsDir); !os.IsNotExist(err) {
		var err error
		createdDirs, err := ioutil.ReadDir(createdDockerConfigsDir)
		if err != nil {
			return false, fmt.Errorf("unable to list created docker configs in %s: %s", createdDockerConfigsDir, err)
		}

		for _, info := range createdDirs {
			if now.Sub(info.ModTime()) > 24*time.Hour {
				return true, nil
			}
		}
	}

	createdWerfConfigRenderFiles := filepath.Join(GetCreatedTmpDirs(), werfConfigRendersServiceDir)
	if _, err := os.Stat(createdWerfConfigRenderFiles); !os.IsNotExist(err) {
		var err error
		createdFiles, err := ioutil.ReadDir(createdWerfConfigRenderFiles)
		if err != nil {
			return false, fmt.Errorf("unable to list created werf config render files in %s: %s", createdWerfConfigRenderFiles, err)
		}

		for _, info := range createdFiles {
			if now.Sub(info.ModTime()) > 24*time.Hour {
				return true, nil
			}
		}
	}

	return false, nil
}

func GC(dryRun bool) error {
	return logboek.LogProcess("Running GC for tmp data", logboek.LogProcessOptions{}, func() error { return gc(dryRun) })
}

func gc(dryRun bool) error {
	projectDirsToRemove := []string{}
	pathsToRemove := []string{}

	if err := gcReleasedProjectDirs(&projectDirsToRemove, &pathsToRemove); err != nil {
		return err
	}

	if err := gcCreatedProjectDirs(&projectDirsToRemove, &pathsToRemove); err != nil {
		return err
	}

	if err := gcCreatedDockerConfigs(&pathsToRemove); err != nil {
		return err
	}

	if err := gcCreatedWerfConfigRenders(&pathsToRemove); err != nil {
		return err
	}

	var removeErrors []error

	if len(projectDirsToRemove) > 0 {
		for _, projectDirToRemove := range projectDirsToRemove {
			logboek.LogLn(projectDirToRemove)
		}

		if !dryRun {
			if runtime.GOOS == "windows" {
				for _, path := range projectDirsToRemove {
					if err := os.RemoveAll(path); err != nil {
						removeErrors = append(removeErrors, fmt.Errorf("unable to remove tmp project dir %s: %s", path, err))
					}
				}
			} else {
				if err := util.RemoveHostDirsWithLinuxContainer(werf.GetTmpDir(), projectDirsToRemove); err != nil {
					removeErrors = append(removeErrors, fmt.Errorf("unable to remove tmp projects dirs %s: %s", strings.Join(projectDirsToRemove, ", "), err))
				}
			}
		}
	}

	for _, path := range pathsToRemove {
		logboek.LogLn(path)

		if !dryRun {
			err := os.RemoveAll(path)
			if err != nil {
				removeErrors = append(removeErrors, fmt.Errorf("unable to remove path %s: %s", path, err))
			}
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

// Remove all released project dirs
func gcReleasedProjectDirs(projectDirsToRemove, pathsToRemove *[]string) error {
	releasedProjectDirsLinks, err := getLinks(filepath.Join(GetReleasedTmpDirs(), projectsServiceDir))
	if err != nil {
		return fmt.Errorf("unable to get released tmp projects dirs: %s", err)
	}

	dirs, err := readLinks(releasedProjectDirsLinks)
	if err != nil {
		return fmt.Errorf("unable to read links: %s", err)
	}
	*projectDirsToRemove = append(*projectDirsToRemove, dirs...)

	for _, link := range releasedProjectDirsLinks {
		*pathsToRemove = append(*pathsToRemove, link.LinkPath)
	}

	return nil
}

// Remove only these created project dirs, which can be removed
func gcCreatedProjectDirs(projectDirsToRemove, pathsToRemove *[]string) error {
	createdProjectDirsLinks, err := getLinks(filepath.Join(GetCreatedTmpDirs(), projectsServiceDir))
	if err != nil {
		return fmt.Errorf("unable to get created tmp projects dirs: %s", err)
	}

	linksToRemove, err := getCreatedFilesToRemove(createdProjectDirsLinks)
	if err != nil {
		return fmt.Errorf("cannot get created tmp files to remove: %s", err)
	}

	dirs, err := readLinks(linksToRemove)
	if err != nil {
		return fmt.Errorf("unable to read links: %s", err)
	}
	*projectDirsToRemove = append(*projectDirsToRemove, dirs...)

	for _, link := range linksToRemove {
		*pathsToRemove = append(*pathsToRemove, link.LinkPath)
	}

	return nil
}

// Remove only these created docker configs, which can be removed
func gcCreatedDockerConfigs(pathsToRemove *[]string) error {
	createdDockerConfigsLinks, err := getLinks(filepath.Join(GetCreatedTmpDirs(), dockerConfigsServiceDir))
	if err != nil {
		return fmt.Errorf("unable to get created tmp docker configs: %s", err)
	}

	linksToRemove, err := getCreatedFilesToRemove(createdDockerConfigsLinks)
	if err != nil {
		return fmt.Errorf("cannot get created tmp files to remove: %s", err)
	}

	dirs, err := readLinks(linksToRemove)
	if err != nil {
		return fmt.Errorf("unable to read links: %s", err)
	}
	*pathsToRemove = append(*pathsToRemove, dirs...)

	for _, link := range linksToRemove {
		*pathsToRemove = append(*pathsToRemove, link.LinkPath)
	}

	return nil
}

// Remove only these created werf config render files, which can be removed
func gcCreatedWerfConfigRenders(pathsToRemove *[]string) error {
	createdWerfConfigRendersLinks, err := getLinks(filepath.Join(GetCreatedTmpDirs(), werfConfigRendersServiceDir))
	if err != nil {
		return fmt.Errorf("unable to get created tmp werf config render files: %s", err)
	}

	linksToRemove, err := getCreatedFilesToRemove(createdWerfConfigRendersLinks)
	if err != nil {
		return fmt.Errorf("cannot get created tmp files to remove: %s", err)
	}

	dirs, err := readLinks(linksToRemove)
	if err != nil {
		return fmt.Errorf("unable to read links: %s", err)
	}
	*pathsToRemove = append(*pathsToRemove, dirs...)

	for _, link := range linksToRemove {
		*pathsToRemove = append(*pathsToRemove, link.LinkPath)
	}

	return nil
}

type LinkDesc struct {
	FileInfo os.FileInfo
	LinkPath string
}

func getLinks(dir string) ([]*LinkDesc, error) {
	var res []*LinkDesc

	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		infos, err := ioutil.ReadDir(dir)
		if err != nil {
			return nil, fmt.Errorf("unable to list files in %s: %s", dir, err)
		}

		for _, info := range infos {
			res = append(res, &LinkDesc{FileInfo: info, LinkPath: filepath.Join(dir, info.Name())})
		}
	}

	return res, nil
}

func readLinks(links []*LinkDesc) ([]string, error) {
	var res []string

	for _, desc := range links {
		origDir, err := os.Readlink(desc.LinkPath)
		if err != nil {
			return nil, fmt.Errorf("unable to read link %s: %s", desc.LinkPath, err)
		}
		res = append(res, origDir)
	}

	return res, nil
}

func getCreatedFilesToRemove(createdFiles []*LinkDesc) ([]*LinkDesc, error) {
	var res []*LinkDesc

	now := time.Now()

	for _, desc := range createdFiles {
		if now.Sub(desc.FileInfo.ModTime()) < 2*time.Hour {
			continue
		}

		res = append(res, desc)
	}

	return res, nil
}
