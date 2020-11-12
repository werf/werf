package util

import (
	"os"
	"path/filepath"
	"strings"
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

func GetRelativeToBaseFilepath(base, path string) string {
	if !filepath.IsAbs(path) {
		if absPath, err := filepath.Abs(path); err != nil {
			panic(err.Error())
		} else {
			path = absPath
		}
	}

	if !filepath.IsAbs(base) {
		if absPath, err := filepath.Abs(base); err != nil {
			panic(err.Error())
		} else {
			base = absPath
		}
	}

	if res, err := filepath.Rel(base, path); err != nil {
		panic(err.Error())
	} else {
		return res
	}
}

func IsSubpathOfBasePath(basePath, path string) bool {
	basePathParts := SplitPath(basePath)
	pathParts := SplitPath(path)

	if len(basePathParts) > len(pathParts) {
		return false
	}
	for i := range basePathParts {
		if basePathParts[i] != pathParts[i] {
			return false
		}
	}
	return true
}

func SplitPath(path string) []string {
	var parts []string
	var dir = path
	var file string
	for {
		dir, file = filepath.Split(dir)
		parts = append([]string{file}, parts...)
		if dir == "" {
			break
		}
		dir = filepath.Dir(dir)
	}
	return parts
}
