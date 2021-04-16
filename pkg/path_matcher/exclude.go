package path_matcher

import (
	"fmt"
	"sort"
	"strings"

	"github.com/werf/werf/pkg/util"
)

func newExcludePathMatcher(excludeGlobs []string) excludePathMatcher {
	return excludePathMatcher{excludeGlobs: formatPaths(excludeGlobs)}
}

type excludePathMatcher struct {
	excludeGlobs []string
}

func (m excludePathMatcher) IsPathMatched(path string) bool {
	path = formatPath(path)
	if len(m.excludeGlobs) > 0 {
		if isAnyPatternMatched(path, m.excludeGlobs) {
			return false
		}
	}

	return true
}

func (m excludePathMatcher) IsDirOrSubmodulePathMatched(path string) bool {
	return m.IsPathMatched(path) || m.ShouldGoThrough(path)
}

func (m excludePathMatcher) ShouldGoThrough(path string) bool {
	return m.shouldGoThrough(formatPath(path))
}

func (m excludePathMatcher) shouldGoThrough(path string) bool {
	if len(m.excludeGlobs) == 0 {
		return false
	}

	if hasUniversalGlob(m.excludeGlobs) {
		return false
	}

	if path == "" {
		return true
	}

	pathParts := util.SplitFilepath(path)
	globsInProgress := m.excludeGlobs
	var matchedGlobs []string
	for _, pathPart := range pathParts {
		if len(globsInProgress) == 0 {
			break
		}

		globsInProgress, matchedGlobs = matchGlobs(pathPart, globsInProgress)
	}

	// the path partially matched any globe and not completely matched any other globe
	if len(matchedGlobs) == 0 && len(globsInProgress) != 0 && !hasUniversalGlob(globsInProgress) {
		return true
	}

	return false
}

func (m excludePathMatcher) ID() string {
	if len(m.excludeGlobs) == 0 {
		return ""
	}

	sort.Strings(m.excludeGlobs)

	var args []string
	args = append(args, "exclude")
	args = append(args, m.excludeGlobs...)
	return util.Sha256Hash(strings.Join(args, ":::"))
}

func (m excludePathMatcher) String() string {
	return fmt.Sprintf("excludeGlobs=%v", m.excludeGlobs)
}
