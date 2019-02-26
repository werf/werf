package util

import (
	"os"
)

// FileExists returns true if path exists
func FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func DirExists(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false, nil
	}

	return fileInfo.IsDir(), nil
}
