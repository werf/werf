package path_matcher

type PathMatcher interface {
	MatchPath(string) bool
	ProcessDirOrSubmodulePath(string) (bool, bool)
	TrimFileBasePath(string) string
	BasePath() string
	String() string
}
