package true_git

type PathFilter interface {
	MatchPath(path string) bool
	TrimFileBasePath(path string) string
	BasePath() string
	String() string
}
