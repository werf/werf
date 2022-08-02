package tmp_manager

import (
	"context"
	"os"
	"path/filepath"
)

func CreateProjectDir(ctx context.Context) (string, error) {
	newDir, err := newTmpDir(ProjectDirPrefix)
	if err != nil {
		return "", err
	}

	if err := registerCreatedPath(newDir, filepath.Join(GetCreatedTmpDirs(), projectsServiceDir)); err != nil {
		os.RemoveAll(newDir)
		return "", err
	}

	return newDir, nil
}

func ReleaseProjectDir(dir string) error {
	return releasePath(dir, filepath.Join(GetCreatedTmpDirs(), projectsServiceDir), filepath.Join(GetReleasedTmpDirs(), projectsServiceDir))
}
