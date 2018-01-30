package config

type RawGitExport struct {
	RawExportBase `yaml:",inline"`
}

func (c *RawGitExport) ToDirective() (gitExport *GitExport, err error) {
	gitExport = &GitExport{}

	if exportBase, err := c.RawExportBase.ToDirective(); err != nil {
		return nil, err
	} else {
		gitExport.ExportBase = exportBase
	}

	gitExport.Raw = c

	if err := c.ValidateDirective(gitExport); err != nil {
		return nil, err
	}

	return gitExport, nil
}

func (c *RawGitExport) ValidateDirective(gitExport *GitExport) (err error) {
	if err := gitExport.Validate(); err != nil {
		return err
	}

	return nil
}
