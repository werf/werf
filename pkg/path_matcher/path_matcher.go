package path_matcher

type PathMatcher interface {
	MatchPath(string) bool
	ProcessDirOrSubmodulePath(string) (bool, bool)
	TrimFileBaseFilepath(string) string
	BaseFilepath() string
	String() string
}
