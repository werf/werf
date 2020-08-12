package tmp_manager

import (
	"context"
	"fmt"
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

	shouldRunGC, err := checkShouldRunGC()
	if err != nil {
		return "", err
	}

	if shouldRunGC {
		err := runGC(ctx)
		if err != nil {
			return "", fmt.Errorf("tmp manager GC failed: %s", err)
		}
	}

	return newFile, nil
}
