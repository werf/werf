package config

type GitExportBase struct {
	*GitExport
	StageDependencies *StageDependencies

	raw *rawGit
}

func (c *GitExportBase) validate() error {
	return nil
}
