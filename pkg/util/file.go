package util

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// FileExists returns true if path exists
func FileExists(p string) (bool, error) {
	_, exist, err := fileExists(p)
	if err != nil {
		return false, fmt.Errorf("unable to check if %q exists: %w", p, err)
	}

	return exist, nil
}

func RegularFileExists(path string) (bool, error) {
	fileInfo, exist, err := fileExists(path)
	if err != nil {
		return false, fmt.Errorf("unable to check if %q is a regular file: %w", path, err)
	}

	regularFileExist := exist && fileInfo.Mode().IsRegular()
	return regularFileExist, nil
}

func DirExists(path string) (bool, error) {
	fileInfo, exist, err := fileExists(path)
	if err != nil {
		return false, fmt.Errorf("unable to check if %q is a directory: %w", path, err)
	}

	dirExist := exist && fileInfo.IsDir()
	return dirExist, nil
}

func fileExists(path string) (os.FileInfo, bool, error) {
	var err error
	path, err = ExpandPath(path)
	if err != nil {
		return nil, false, fmt.Errorf("unable to expand path %q: %w", path, err)
	}

	fileInfo, err := os.Lstat(path)
	if err != nil {
		if isNotExistError(err) {
			return nil, false, nil
		}

		return nil, false, err
	}

	return fileInfo, true, nil
}

func isNotExistError(err error) bool {
	return os.IsNotExist(err) || IsNotADirectoryError(err)
}

func IsNotADirectoryError(err error) bool {
	return strings.HasSuffix(err.Error(), "not a directory")
}

func GetRelativeToBaseFilepath(base, path string) string {
	path = GetAbsoluteFilepath(path)
	base = GetAbsoluteFilepath(base)

	res, err := filepath.Rel(base, path)
	if err != nil {
		panic(err.Error())
	}

	return res
}

func GetAbsoluteFilepath(absOrRelPath string) string {
	absPath, err := filepath.Abs(absOrRelPath)
	if err != nil {
		panic(err.Error())
	}

	return absPath
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
