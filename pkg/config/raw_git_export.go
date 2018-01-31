package config

type RawGitExport struct {
	RawExportBase `yaml:",inline"`

	RawOrigin RawOrigin `yaml:"-"` // parent
}

func (c *RawGitExport) InlinedIntoRaw(RawOrigin RawOrigin) {
	c.RawOrigin = RawOrigin
	c.RawExportBase.InlinedIntoRaw(RawOrigin)
}

func NewRawGitExport() RawGitExport {
	rawGitExport := RawGitExport{}
	rawGitExport.RawExportBase = RawExportBase{}
	rawGitExport.RawExportBase.Add = "/"
	return rawGitExport
}

func (c *RawGitExport) ToDirective() (gitExport *GitExport, err error) {
	gitExport = &GitExport{}

	if exportBase, err := c.RawExportBase.ToDirective(); err != nil {
		return nil, err
	} else {
		gitExport.ExportBase = exportBase
	}

	gitExport.Raw = c

	if err := gitExport.Validate(); err != nil {
		return nil, err
	}

	return gitExport, nil
}
