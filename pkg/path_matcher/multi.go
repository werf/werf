package path_matcher

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
)

func NewMultiPathMatcher(matchers ...PathMatcher) PathMatcher {
	if len(matchers) == 0 {
		matchers = append(matchers, NewTruePathMatcher())
	}

	return MultiPathMatcher{matchers: matchers}
}

type MultiPathMatcher struct {
	matchers []PathMatcher
}

func (m MultiPathMatcher) IsDirOrSubmodulePathMatched(path string) bool {
	return m.IsPathMatched(path) || m.ShouldGoThrough(path)
}

func (m MultiPathMatcher) IsPathMatched(path string) bool {
	for _, matcher := range m.matchers {
		if !matcher.IsPathMatched(path) {
			return false
		}
	}

	return true
}

// ShouldGoThrough returns true if the ShouldGoThrough method of at least one matcher returns true and the path partially or completely matched by others (IsDirOrSubmodulePathMatched returns true)
func (m MultiPathMatcher) ShouldGoThrough(path string) bool {
	var shouldGoThrough bool
	for _, matcher := range m.matchers {
		if matcher.ShouldGoThrough(path) {
			shouldGoThrough = true
		} else if !matcher.IsPathMatched(path) {
			return false
		}
	}

	return shouldGoThrough
}

func (m MultiPathMatcher) ID() string {
	h := sha256.New()

	var ids []string
	for _, matcher := range m.matchers {
		ids = append(ids, matcher.ID())
	}

	sort.Strings(ids)
	h.Write([]byte(fmt.Sprint(ids)))

	return fmt.Sprintf("%x", h.Sum(nil))
}

func (m MultiPathMatcher) String() string {
	var result []string
	for _, matcher := range m.matchers {
		result = append(result, matcher.String())
	}

	return fmt.Sprintf("{ %s }", strings.Join(result, " && "))
}
