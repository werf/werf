package config

type GitExport struct {
	*ExportBase

	Raw *RawGitExport
}

func (c *GitExport) Validate() error {
	return nil
}
