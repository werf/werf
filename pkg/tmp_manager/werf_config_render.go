package tmp_manager

import (
	"os"
	"path/filepath"
)

func CreateWerfConfigRender(dir string) (string, error) {
	newFile, err := newTmpFile(dir, WerfConfigRenderPrefix)
	if err != nil {
		return "", err
	}

	if err := registerCreatedPath(newFile, filepath.Join(GetCreatedTmpDirs(), werfConfigRendersServiceDir)); err != nil {
		os.RemoveAll(newFile)
		return "", err
	}

	return newFile, nil
}
