package util

import (
	"os"
	"strings"
	"reflect"
)

// FileExists returns true if path exists
func FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if isNotExistError(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func DirExists(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		if isNotExistError(err) {
			return false, nil
		}

		return false, err
	}

	return fileInfo.IsDir(), nil
}

func isNotExistError(err error) bool {
	return os.IsNotExist(err) || IsNotADirectoryError(err)
}

func IsNotADirectoryError(err error) bool {
	return strings.HasSuffix(err.Error(), "not a directory")
}

func IsSubpathOfBasePath(basePath, path string) bool {
	basePathParts := SplitFilepath(basePath)
	pathParts := SplitFilepath(path)

	if len(basePathParts) > len(pathParts) {
		return false
	}

	if reflect.DeepEqual(basePathParts, pathParts) {
		return false
	}

	for ind := range basePathParts {
		if basePathParts[ind] == "" {
			continue
		}

		if basePathParts[ind] != pathParts[ind] {
			return false
		}
	}

	return true
}
