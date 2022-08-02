package tmp_manager

import (
	"context"
	"os"
	"path/filepath"
)

func CreateWerfConfigRender(ctx context.Context) (string, error) {
	newFile, err := newTmpFile(WerfConfigRenderPrefix)
	if err != nil {
		return "", err
	}

	if err := registerCreatedPath(newFile, filepath.Join(GetCreatedTmpDirs(), werfConfigRendersServiceDir)); err != nil {
		os.RemoveAll(newFile)
		return "", err
	}

	return newFile, nil
}
