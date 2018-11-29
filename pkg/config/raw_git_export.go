package config

type rawGitExport struct {
	rawExportBase `yaml:",inline"`

	rawOrigin rawOrigin `yaml:"-"` // parent
}

func (c *rawGitExport) inlinedIntoRaw(rawOrigin rawOrigin) {
	c.rawOrigin = rawOrigin
	c.rawExportBase.inlinedIntoRaw(rawOrigin)
}

func newRawGitExport() rawGitExport {
	rawGitExport := rawGitExport{}
	rawGitExport.rawExportBase = rawExportBase{}
	rawGitExport.rawExportBase.Add = "/"
	return rawGitExport
}

func (c *rawGitExport) toDirective() (gitExport *GitExport, err error) {
	gitExport = &GitExport{}

	if exportBase, err := c.rawExportBase.toDirective(); err != nil {
		return nil, err
	} else {
		gitExport.ExportBase = exportBase
	}

	gitExport.raw = c

	if err := gitExport.validate(); err != nil {
		return nil, err
	}

	return gitExport, nil
}
