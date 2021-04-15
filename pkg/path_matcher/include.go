package path_matcher

import (
	"fmt"
	"sort"
	"strings"

	"github.com/werf/werf/pkg/util"
)

func newIncludePathMatcher(includePaths []string) includePathMatcher {
	return includePathMatcher{includeGlobs: formatPaths(includePaths)}
}

type includePathMatcher struct {
	includeGlobs []string
}

func (m includePathMatcher) IsPathMatched(path string) bool {
	path = formatPath(path)
	if len(m.includeGlobs) > 0 {
		if !isAnyPatternMatched(path, m.includeGlobs) {
			return false
		}
	}

	return true
}

func (m includePathMatcher) IsDirOrSubmodulePathMatched(path string) bool {
	return m.IsPathMatched(path) || m.ShouldGoThrough(path)
}

func (m includePathMatcher) ShouldGoThrough(path string) bool {
	return m.shouldGoThrough(formatPath(path))
}

func (m includePathMatcher) shouldGoThrough(path string) bool {
	if len(m.includeGlobs) == 0 {
		return false
	}

	if hasUniversalGlob(m.includeGlobs) {
		return false
	}

	if path == "" {
		return true
	}

	pathParts := util.SplitFilepath(path)
	globsInProgress := m.includeGlobs
	for _, pathPart := range pathParts {
		if len(globsInProgress) == 0 {
			break
		}

		if len(globsInProgress) != 0 {
			globsInProgress, _ = matchGlobs(pathPart, globsInProgress)
		}
	}

	// the path partially matched any globe
	if len(globsInProgress) != 0 && !hasUniversalGlob(globsInProgress) {
		return true
	}

	return false
}

func (m includePathMatcher) ID() string {
	if len(m.includeGlobs) == 0 {
		return ""
	}

	sort.Strings(m.includeGlobs)

	var args []string
	args = append(args, "include")
	args = append(args, m.includeGlobs...)
	return util.Sha256Hash(strings.Join(args, ":::"))
}

func (m includePathMatcher) String() string {
	return fmt.Sprintf("{ includeGlobs=%v }", m.includeGlobs)
}
