package git_repo

type GitRepo interface {
	HeadCommit() (string, error)
	LatestBranchCommit(branch string) (string, error)
	LatestTagCommit(tag string) (string, error)
}
