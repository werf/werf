package path_matcher

import (
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar"

	"github.com/werf/werf/pkg/util"
)

func matchGlobs(pathPart string, globs []string) (inProgressGlobs, matchedGlobs []string) {
	for _, glob := range globs {
		inProgressGlob, matchedGlob := matchGlob(pathPart, glob)
		if inProgressGlob != "" {
			inProgressGlobs = append(inProgressGlobs, inProgressGlob)
		} else if matchedGlob != "" {
			matchedGlobs = append(matchedGlobs, matchedGlob)
		}
	}

	return
}

func matchGlob(pathPart, glob string) (inProgressGlob, matchedGlob string) {
	globParts := util.SplitFilepath(glob)
	isMatched, err := doublestar.PathMatch(globParts[0], pathPart)
	if err != nil {
		panic(err)
	}
	if !isMatched {
		return "", ""
	}
	if strings.Contains(globParts[0], "**") {
		return glob, ""
	}
	if len(globParts) > 1 {
		return filepath.Join(globParts[1:]...), ""
	}
	return "", glob
}

func hasUniversalGlob(globs []string) bool {
	for _, glob := range globs {
		if glob == "." {
			return true
		}

		if util.SafeTrimGlobsAndSlashesFromFilepath(glob) == "" {
			return true
		}
	}

	return false
}

func isAnyPatternMatched(path string, patterns []string) bool {
	for _, pattern := range patterns {
		if isPathMatched(path, pattern) {
			return true
		}
	}

	return false
}

func isPathMatched(filePath, glob string) bool {
	// The glob as-is.
	// The glob without asterisks on the right (path/*/dir/**/*, path/*/dir/**, path/*/dir/*/*/** -> path/*/dir).
	// The previous glob with the universal part `**/*` (path/*/dir/**/*).
	for _, g := range []string{
		glob,
		util.SafeTrimGlobsAndSlashesFromFilepath(glob),
		filepath.Join(util.SafeTrimGlobsAndSlashesFromFilepath(glob), "**", "*"),
	} {
		matched, err := doublestar.PathMatch(g, filePath)
		if err != nil {
			panic(err)
		}

		if matched {
			return true
		}
	}

	return false
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
