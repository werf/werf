package path_matcher

func NewFalsePathMatcher() PathMatcher {
	return FalsePathMatcher{}
}

// FalsePathMatcher always returns false when matching paths
type FalsePathMatcher struct{}

func (m FalsePathMatcher) IsDirOrSubmodulePathMatched(_ string) bool {
	return false
}

func (m FalsePathMatcher) IsPathMatched(_ string) bool {
	return false
}

func (m FalsePathMatcher) ShouldGoThrough(_ string) bool {
	return false
}

func (m FalsePathMatcher) ID() string {
	return m.String()
}

func (m FalsePathMatcher) String() string {
	return "{ false }"
}
