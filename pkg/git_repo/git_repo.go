package git_repo

type GitRepo interface {
	String() string
	HeadCommit() (string, error)
	LatestBranchCommit(branch string) (string, error)
	LatestTagCommit(tag string) (string, error)
}
