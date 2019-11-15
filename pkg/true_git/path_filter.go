package true_git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar"
)

type PathFilter struct {
	// BasePath can be path to the directory or a single file
	BasePath                   string
	IncludePaths, ExcludePaths []string
}

func (f *PathFilter) IsFilePathValid(filePath string) bool {
	return isFilePathValid(filePath, f.BasePath, f.IncludePaths, f.ExcludePaths)
}

func (f *PathFilter) TrimFileBasePath(filePath string) string {
	return trimFileBasePath(filePath, f.BasePath)
}

func (f *PathFilter) String() string {
	return fmt.Sprintf("BasePath=`%s`, IncludePaths=%v, ExcludePaths=%v", f.BasePath, f.IncludePaths, f.ExcludePaths)
}

func isFilePathValid(filePath, basePath string, includePaths, excludePaths []string) bool {
	if !isRel(filePath, basePath) {
		return false
	}

	relFilePath := rel(filePath, basePath)

	if len(includePaths) > 0 {
		if !isPathMatchedOneOfPatterns(relFilePath, includePaths) {
			return false
		}
	}

	if len(excludePaths) > 0 {
		if isPathMatchedOneOfPatterns(relFilePath, excludePaths) {
			return false
		}
	}

	return true
}

func isRel(targetPath, basePath string) bool {
	if basePath == "" {
		return true
	}
	return strings.HasPrefix(targetPath+string(os.PathSeparator), basePath+string(os.PathSeparator))
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

func isPathMatchedPattern(filePath, pattern string) bool {
	matched, err := doublestar.PathMatch(pattern, filePath)
	if err != nil {
		panic(err)
	}
	if matched {
		return true
	}

	matched, err = doublestar.PathMatch(filepath.Join(pattern, "**", "*"), filePath)
	if err != nil {
		panic(err)
	}
	if matched {
		return true
	}

	return false
}

func isPathMatchedOneOfPatterns(path string, patterns []string) bool {
	for _, pattern := range patterns {
		if isRel(path, pattern) {
			return true
		}

		if isPathMatchedPattern(path, pattern) {
			return true
		}
	}

	return false
}

func trimFileBasePath(filePath, basePath string) string {
	if filePath == basePath {
		// filePath path is always a path to a file, not a directory.
		// Thus if BasePath is equal filePath, then BasePath is a path to the file.
		// Return file name in this case by convention.
		return filepath.Base(filePath)
	}

	return rel(filePath, basePath)
}
