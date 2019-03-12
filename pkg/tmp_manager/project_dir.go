package tmp_manager

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/stapel"
	"github.com/flant/werf/pkg/werf"
)

func CreateProjectDir() (string, error) {
	newDir, err := newTmpDir(ProjectDirPrefix)
	if err != nil {
		return "", err
	}

	if err := registerCreatedPath(newDir, filepath.Join(GetCreatedTmpDirs(), projectsServiceDir)); err != nil {
		os.RemoveAll(newDir)
		return "", err
	}

	shouldRunGC, err := checkShouldRunGC()
	if err != nil {
		return "", err
	}

	if shouldRunGC {
		err := runGC()
		if err != nil {
			return "", fmt.Errorf("tmp manager GC failed: %s", err)
		}
	}

	return newDir, nil
}

func ReleaseProjectDir(dir string) error {
	return releasePath(dir, filepath.Join(GetCreatedTmpDirs(), projectsServiceDir), filepath.Join(GetReleasedTmpDirs(), projectsServiceDir))
}

func removeProjectDirs(dirs []string) error {
	args := []string{
		"--rm",
		"--volume", fmt.Sprintf("%s:%s", werf.GetTmpDir(), werf.GetTmpDir()),
		stapel.ImageName(),
		stapel.RmBinPath(), "-rf",
	}

	args = append(args, dirs...)

	return docker.CliRun(args...)
}
