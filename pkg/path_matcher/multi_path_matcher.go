package path_matcher

import (
	"strings"
)

func NewMultiPathMatcher(PathMatchers ...PathMatcher) PathMatcher {
	if len(PathMatchers) == 0 {
		panic("the multi path matcher cannot be initialized without any matcher")
	}

	return &MultiPathMatcher{PathMatchers: PathMatchers}
}

type MultiPathMatcher struct {
	PathMatchers []PathMatcher
}

func (m *MultiPathMatcher) MatchPath(path string) bool {
	for _, matcher := range m.PathMatchers {
		if !matcher.MatchPath(path) {
			return false
		}
	}

	return true
}

func (m *MultiPathMatcher) ProcessDirOrSubmodulePath(path string) (bool, bool) {
	var isMatchedResults, shouldGoThroughResults []bool
	for _, matcher := range m.PathMatchers {
		isMatched, shouldGoThrough := matcher.ProcessDirOrSubmodulePath(path)
		isMatchedResults = append(isMatchedResults, isMatched)
		shouldGoThroughResults = append(shouldGoThroughResults, shouldGoThrough)
	}

	var isMatched bool
	var shouldGoThrough bool
	for _, isMatchedResult := range isMatchedResults {
		if !isMatchedResult {
			isMatched = false
			break
		}

		isMatched = true
	}

	for _, isMatchedResult := range shouldGoThroughResults {
		if !isMatchedResult {
			shouldGoThrough = false
			break
		}

		shouldGoThrough = true
	}

	return isMatched, shouldGoThrough
}

func (m *MultiPathMatcher) TrimFileBaseFilepath(path string) string {
	return m.PathMatchers[0].TrimFileBaseFilepath(path)
}

func (m *MultiPathMatcher) BaseFilepath() string {
	return m.PathMatchers[0].BaseFilepath()
}

func (m *MultiPathMatcher) String() string {
	var result []string
	for _, matcher := range m.PathMatchers {
		result = append(result, matcher.String())
	}

	return strings.Join(result, "; ")
}
