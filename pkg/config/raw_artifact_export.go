package config

type RawArtifactExport struct {
	RawExportBase `yaml:",inline"`

	RawOrigin RawOrigin `yaml:"-"` // parent
}

func (c *RawArtifactExport) InlinedIntoRaw(RawOrigin RawOrigin) {
	c.RawOrigin = RawOrigin
	c.RawExportBase.InlinedIntoRaw(RawOrigin)
}

func (c *RawArtifactExport) ToDirective() (artifactExport *ArtifactExport, err error) {
	artifactExport = &ArtifactExport{}

	if exportBase, err := c.RawExportBase.ToDirective(); err != nil {
		return nil, err
	} else {
		artifactExport.ExportBase = exportBase
	}

	artifactExport.Raw = c

	if err := artifactExport.Validate(); err != nil {
		return nil, err
	}

	return artifactExport, nil
}
