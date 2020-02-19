package true_git

type PathMatcher interface {
	MatchPath(path string) bool
	TrimFileBasePath(path string) string
	BasePath() string
	String() string
}
