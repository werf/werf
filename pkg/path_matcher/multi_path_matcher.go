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

func (f *MultiPathMatcher) IsDirOrSubmodulePathMatched(path string) bool {
	return f.IsPathMatched(path) || f.ShouldGoThrough(path)
}

func (m *MultiPathMatcher) IsPathMatched(path string) bool {
	for _, matcher := range m.PathMatchers {
		if !matcher.IsPathMatched(path) {
			return false
		}
	}

	return true
}

func (m *MultiPathMatcher) ShouldGoThrough(path string) bool {
	for _, matcher := range m.PathMatchers {
		if !matcher.ShouldGoThrough(path) {
			return false
		}
	}

	return true
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

	return strings.Join(result, " && ")
}
