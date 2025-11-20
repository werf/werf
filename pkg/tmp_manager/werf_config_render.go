package tmp_manager

import (
	"context"
	"fmt"
	"path/filepath"
)

func CreateWerfConfigRender(ctx context.Context) (string, error) {
	newFile, err := newTmpFile(werfConfigRenderPrefix)
	if err != nil {
		return "", err
	}

	if err := registrator.queueRegistration(ctx, newFile, filepath.Join(getCreatedTmpDirs(), werfConfigRendersServiceDir)); err != nil {
		return "", fmt.Errorf("unable to queue GC registration: %w", err)
	}

	return newFile, nil
}
