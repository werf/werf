package tmp_manager

import (
	"context"
	"fmt"
	"path/filepath"
)

func CreateProjectDir(ctx context.Context) (string, error) {
	newDir, err := newTmpDir(projectDirPrefix)
	if err != nil {
		return "", err
	}

	if err := registrator.queueRegistration(ctx, newDir, filepath.Join(getCreatedTmpDirs(), projectsServiceDir)); err != nil {
		return "", fmt.Errorf("unable to queue GC registration: %w", err)
	}
	if err := registrator.queueRegistration(ctx, newDir, filepath.Join(getReleasedTmpDirs(), projectsServiceDir)); err != nil {
		return "", fmt.Errorf("unable to queue GC registration: %w", err)
	}

	return newDir, nil
}
