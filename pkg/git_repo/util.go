package git_repo

import (
	"os"
	"path/filepath"
)

func renameFile(fromPath, toPath string) error {
	var err error

	err = os.MkdirAll(filepath.Dir(toPath), os.ModePerm)
	if err != nil {
		return err
	}

	err = os.Rename(fromPath, toPath)
	if err != nil {
		return err
	}

	return nil
}
