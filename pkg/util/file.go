package util

import "os"

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func DirExists(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false, nil
	}

	return fileInfo.IsDir(), nil
}
