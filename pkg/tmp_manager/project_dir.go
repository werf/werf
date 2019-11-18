package tmp_manager

import (
	"fmt"
	"os"
	"path/filepath"
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
