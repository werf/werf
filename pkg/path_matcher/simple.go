package path_matcher

import (
	"crypto/sha256"
	"fmt"
	"sort"

	"github.com/werf/werf/pkg/util"
)

func NewSimplePathMatcher(basePath string, paths []string) *SimplePathMatcher {
	return &SimplePathMatcher{
		basePathMatcher: basePathMatcher{basePath: formatPath(basePath)},
		paths:           formatPaths(paths),
	}
}

type SimplePathMatcher struct {
	basePathMatcher
	paths []string
}

func (f *SimplePathMatcher) BaseFilepath() string {
	return f.basePath
}

func (f *SimplePathMatcher) Paths() []string {
	return f.paths
}

func (f *SimplePathMatcher) ID() string {
	h := sha256.New()

	{ // basePath
		h.Write([]byte(f.basePath))
	}

	{ // paths
		if len(f.paths) != 0 {
			sort.Strings(f.paths)
			h.Write([]byte(fmt.Sprint(f.paths)))
		}
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

func (f *SimplePathMatcher) String() string {
	return fmt.Sprintf("basePath=`%s`, paths=%v", f.basePath, f.paths)
}

func (f *SimplePathMatcher) IsDirOrSubmodulePathMatched(path string) bool {
	return f.IsPathMatched(path) || f.ShouldGoThrough(path)
}

func (f *SimplePathMatcher) IsPathMatched(path string) bool {
	path = formatPath(path)

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

func (f *SimplePathMatcher) ShouldGoThrough(path string) bool {
	return f.shouldGoThrough(formatPath(path))
}

func (f *SimplePathMatcher) shouldGoThrough(path string) bool {
	isBasePathRelativeToPath := isSubDirOf(f.basePath, path)
	isPathRelativeToBasePath := isSubDirOf(path, f.basePath)

	if isPathRelativeToBasePath || path == f.basePath {
		if len(f.paths) == 0 {
			return false
		} else if hasUniversalGlob(f.paths) {
			return false
		} else if path == f.basePath {
			return true
		}

		return f.shouldGoThroughDetailedCheck(path)
	} else if isBasePathRelativeToPath {
		return true
	} else { // path is not relative to basePath
		return false
	}
}

func (f *SimplePathMatcher) shouldGoThroughDetailedCheck(path string) bool {
	relPath := rel(path, f.basePath)
	relPathParts := util.SplitFilepath(relPath)
	inProgressPaths := f.paths[:]
	var matchedIncludePaths []string

	for _, pathPart := range relPathParts {
		if len(inProgressPaths) == 0 {
			break
		}

		if len(inProgressPaths) != 0 {
			inProgressPaths, matchedIncludePaths = matchGlobs(pathPart, inProgressPaths)
			if len(inProgressPaths) == 0 && len(matchedIncludePaths) == 0 {
				return false
			}
		}
	}

	if len(inProgressPaths) != 0 {
		return !hasUniversalGlob(inProgressPaths)
	}

	return false
}
