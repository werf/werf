package config

type GitLocalExport struct {
	*GitExportBase
	Lfs bool

	raw *rawGit
}

func (c *GitLocalExport) validate() error {
	return nil
}
