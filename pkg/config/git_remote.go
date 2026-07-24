package config

type GitRemote struct {
	*GitRemoteExport
	Name         string
	Url          string
	RepoCacheKey string

	raw *rawGit
}

func (c *GitRemote) GetRaw() interface{} {
	return c.raw
}

func (c *GitRemote) validate() error {
	return nil
}
