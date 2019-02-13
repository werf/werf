package tmp_manager

import (
	"fmt"
	"os"
	"path/filepath"
)

func CreateWerfConfigRender() (string, error) {
	newFile, err := newTmpFile("werf-config-render-")
	if err != nil {
		return "", err
	}

	if err := registerCreatedPath(newFile, filepath.Join(GetCreatedTmpDirs(), werfConfigRendersDir)); err != nil {
		os.RemoveAll(newFile)
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

	return newFile, nil
}
