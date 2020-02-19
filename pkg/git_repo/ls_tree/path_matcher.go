package ls_tree

type PathMatcher interface {
	MatchPath(path string) bool
	ProcessDirOrSubmodulePath(path string) (bool, bool)
}
