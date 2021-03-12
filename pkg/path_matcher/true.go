package path_matcher

func NewTruePathMatcher() PathMatcher {
	return &TruePathMatcher{}
}

// TruePathMatcher always returns true when matching paths
type TruePathMatcher struct{}

func (m TruePathMatcher) IsDirOrSubmodulePathMatched(_ string) bool {
	return true
}

func (m TruePathMatcher) IsPathMatched(_ string) bool {
	return true
}

func (m TruePathMatcher) ShouldGoThrough(_ string) bool {
	return false
}

func (m TruePathMatcher) ID() string {
	return m.String()
}

func (m TruePathMatcher) String() string {
	return "{ true }"
}
