package path_matcher

import (
	"crypto/sha256"
	"fmt"
	"sort"
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

func (m *MultiPathMatcher) IsDirOrSubmodulePathMatched(path string) bool {
	return m.IsPathMatched(path) || m.ShouldGoThrough(path)
}

func (m *MultiPathMatcher) IsPathMatched(path string) bool {
	for _, matcher := range m.PathMatchers {
		if !matcher.IsPathMatched(path) {
			return false
		}
	}

	return true
}

// ShouldGoThrough returns true if the ShouldGoThrough method of at least one matcher returns true and the path partially or completely matched by others (IsDirOrSubmodulePathMatched returns true)
func (m *MultiPathMatcher) ShouldGoThrough(path string) bool {
	var shouldGoThrough bool
	for _, matcher := range m.PathMatchers {
		if matcher.ShouldGoThrough(path) {
			shouldGoThrough = true
		} else if !matcher.IsPathMatched(path) {
			return false
		}
	}

	return shouldGoThrough
}

func (m *MultiPathMatcher) TrimFileBaseFilepath(path string) string {
	return m.PathMatchers[0].TrimFileBaseFilepath(path)
}

func (m *MultiPathMatcher) BaseFilepath() string {
	return m.PathMatchers[0].BaseFilepath()
}

func (m *MultiPathMatcher) ID() string {
	h := sha256.New()

	var ids []string
	for _, matcher := range m.PathMatchers {
		ids = append(ids, matcher.ID())
	}

	sort.Strings(ids)
	h.Write([]byte(fmt.Sprint(ids)))

	return fmt.Sprintf("%x", h.Sum(nil))
}

func (m *MultiPathMatcher) String() string {
	var result []string
	for _, matcher := range m.PathMatchers {
		result = append(result, matcher.String())
	}

	return strings.Join(result, " && ")
}
