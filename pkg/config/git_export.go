package config

type GitExport struct {
	*ExportBase

	raw *rawGitExport
}

func (c *GitExport) validate() error {
	return nil
}
