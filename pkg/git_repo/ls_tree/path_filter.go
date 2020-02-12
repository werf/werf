package ls_tree

type PathFilter interface {
	MatchPath(path string) bool
	ProcessDirOrSubmodulePath(path string) (bool, bool)
}
