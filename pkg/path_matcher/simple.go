package path_matcher

import (
	"fmt"

	"github.com/werf/werf/pkg/util"
)

func NewSimplePathMatcher(basePath string, paths []string, greedySearch bool) *SimplePathMatcher {
	return &SimplePathMatcher{
		basePathMatcher:  basePathMatcher{basePath: formatPath(basePath)},
		paths:            formatPaths(paths),
		isGreedySearchOn: greedySearch,
	}
}

type SimplePathMatcher struct {
	basePathMatcher
	paths            []string
	isGreedySearchOn bool
}

func (f *SimplePathMatcher) BaseFilepath() string {
	return f.basePath
}

func (f *SimplePathMatcher) Paths() []string {
	return f.paths
}

func (f *SimplePathMatcher) String() string {
	return fmt.Sprintf("basePath=`%s`, paths=%v, greedySearch=%v", f.basePath, f.paths, f.isGreedySearchOn)
}

func (f *SimplePathMatcher) MatchPath(path string) bool {
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

func (f *SimplePathMatcher) ProcessDirOrSubmodulePath(path string) (bool, bool) {
	isMatched, shouldGoThrough := f.processDirOrSubmodulePath(formatPath(path))
	if f.isGreedySearchOn {
		return false, isMatched || shouldGoThrough
	} else {
		return isMatched, shouldGoThrough
	}
}

func (f *SimplePathMatcher) processDirOrSubmodulePath(path string) (bool, bool) {
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
