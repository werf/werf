package cleaning

type GitRepo interface {
	IsCommitExists(commit string) (bool, error)
	TagsList() ([]string, error)
	RemoteBranchesList() ([]string, error)
}
