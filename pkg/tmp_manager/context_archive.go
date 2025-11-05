package tmp_manager

import (
	"context"
	"fmt"
	"path/filepath"
)

func CreateContextArchivePath(ctx context.Context) (string, error) {
	newFile, err := newTmpFile(contextArchivePrefix)
	if err != nil {
		return "", err
	}

	if err := registrator.queueRegistration(ctx, newFile, filepath.Join(getCreatedTmpDirs(), contextArchivesDir)); err != nil {
		return "", fmt.Errorf("unable to queue GC registration: %w", err)
	}

	return newFile, nil
}
