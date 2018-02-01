package config

type GitExportBase struct {
	*GitExport
	StageDependencies *StageDependencies

	Raw *RawGit
}

func (c *GitExportBase) Validate() error {
	return nil
}
