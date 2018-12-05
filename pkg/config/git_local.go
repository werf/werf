package config

type GitLocal struct {
	*GitLocalExport
	As string

	raw *rawGit
}

func (c *GitLocal) validate() error {
	return nil
}
