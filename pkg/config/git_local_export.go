package config

type GitLocalExport struct {
	*GitExportBase

	raw *rawGit
}

func (c *GitLocalExport) validate() error {
	return nil
}
