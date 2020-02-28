package path_matcher

import (
	"fmt"
	"os"
	"strings"
)

func NewSimplePathMatcher(basePath string, paths []string) *SimplePathMatcher {
	return &SimplePathMatcher{
		basePath: basePath,
		paths:    paths,
	}
}

type SimplePathMatcher struct {
	basePath string
	paths    []string
}

func (f *SimplePathMatcher) BasePath() string {
	return f.basePath
}

func (f *SimplePathMatcher) Paths() []string {
	return f.paths
}

func (f *SimplePathMatcher) String() string {
	return fmt.Sprintf("basePath=`%s`, paths=%v", f.basePath, f.paths)
}

func (f *SimplePathMatcher) MatchPath(path string) bool {
	if !isRel(path, f.basePath) {
		return false
	}

	subPathOrDot := rel(path, f.basePath)

	if len(f.paths) > 0 {
		if !isAnyPatternMatched(subPathOrDot, f.paths) {
			return false
		}
	}

	return true
}

func (f *SimplePathMatcher) ProcessDirOrSubmodulePath(path string) (bool, bool) {
	isBasePathRelativeToPath := isSubDirOf(f.basePath, path)
	isPathRelativeToBasePath := isSubDirOf(path, f.basePath)

	if isPathRelativeToBasePath || path == f.basePath {
		if len(f.paths) == 0 {
			return true, false
		} else if hasUniversalGlob(f.paths) {
			return true, false
		} else if path == f.basePath {
			return false, true
		}
	} else if isBasePathRelativeToPath {
		return false, true
	} else { // path is not relative to basePath
		return false, false
	}

	relPath := rel(path, f.basePath)
	relPathParts := strings.Split(relPath, string(os.PathSeparator))
	inProgressPaths := f.paths[:]
	var matchedIncludePaths []string

	for _, pathPart := range relPathParts {
		if len(inProgressPaths) == 0 {
			break
		}

		if len(inProgressPaths) != 0 {
			inProgressPaths, matchedIncludePaths = matchGlobs(pathPart, inProgressPaths)
			if len(inProgressPaths) == 0 && len(matchedIncludePaths) == 0 {
				return false, false
			}
		}
	}

	if len(inProgressPaths) != 0 {
		if hasUniversalGlob(inProgressPaths) {
			return true, false
		} else {
			return false, true
		}
	} else if len(matchedIncludePaths) != 0 {
		return true, false
	} else {
		return false, false
	}
}
