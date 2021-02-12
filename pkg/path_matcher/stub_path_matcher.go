package path_matcher

func NewStubPathMatcher(basePath string) *StubPathMatcher {
	return &StubPathMatcher{SimplePathMatcher: NewSimplePathMatcher(basePath, nil, true)}
}

// StubPathMatcher returns false when matching paths
type StubPathMatcher struct {
	*SimplePathMatcher
}

func (m *StubPathMatcher) MatchPath(_ string) bool {
	return false
}

func (m *StubPathMatcher) ProcessDirOrSubmodulePath(_ string) (bool, bool) {
	return false, false
}

func (m *StubPathMatcher) String() string {
	return "none"
}
