package config

type GitRemote struct {
	*GitRemoteExport
	As   string
	Name string
	Url  string

	raw *rawGit
}

func (c *GitRemote) validate() error {
	return nil
}
