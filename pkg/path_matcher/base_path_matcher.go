package path_matcher

import (
	"path/filepath"
)

type basePathMatcher struct {
	basePath string
}

func (f *basePathMatcher) TrimFileBaseFilepath(path string) string {
	return trimFileBasePath(formatPath(path), f.basePath)
}

func trimFileBasePath(filePath, basePath string) string {
	if filePath == basePath {
		// filePath path is always a path to a file, not a directory.
		// Thus if basePath is equal filePath, then basePath is a path to the file.
		// Return file name in this case by convention.
		return filepath.Base(filePath)
	}

	return rel(filePath, basePath)
}

func rel(targetPath, basePath string) string {
	if basePath == "" {
		return targetPath
	}

	relPath, err := filepath.Rel(basePath, targetPath)
	if err != nil {
		panic(err)
	}

	return relPath
}

func formatPaths(paths []string) []string {
	var result []string
	for _, path := range paths {
		result = append(result, formatPath(path))
	}

	return result
}

func formatPath(path string) string {
	if path == "" || path == "." {
		return ""
	}

	return filepath.Clean(path)
}
