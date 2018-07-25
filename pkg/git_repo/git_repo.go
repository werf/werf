package git_repo

type GitRepo interface {
	String() string
	HeadCommit() (string, error)
	HeadBranchName() (string, error)
	LatestBranchCommit(branch string) (string, error)
	LatestTagCommit(tag string) (string, error)
	Diff(basePath string, fromCommit, toCommit string, includePaths, excludePaths []string) (string, error)
	IsAnyChanges(basePath string, fromCommit, toCommit string, includePaths, excludePaths []string) (bool, error)
}
