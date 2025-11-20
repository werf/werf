package file_manager

import (
	"fmt"
	"os"
)

func newFileWithPath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("empty path")
	}

	newFile, err := os.Create(path)
	if err != nil {
		return "", err
	}

	path = newFile.Name()

	if err = newFile.Close(); err != nil {
		return "", err
	}

	return path, err
}
