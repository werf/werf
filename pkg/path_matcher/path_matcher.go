package path_matcher

type PathMatcher interface {
	MatchPath(string) bool
	ProcessDirOrSubmodulePath(string) (bool, bool)
	BasePath() string
	String() string
}
