package config

type rawArtifactExport struct {
	rawExportBase `yaml:",inline"`

	rawOrigin rawOrigin `yaml:"-"` // parent
}

func (c *rawArtifactExport) inlinedIntoRaw(rawOrigin rawOrigin) {
	c.rawOrigin = rawOrigin
	c.rawExportBase.inlinedIntoRaw(rawOrigin)
}

func (c *rawArtifactExport) toDirective() (artifactExport *ArtifactExport, err error) {
	artifactExport = &ArtifactExport{}

	if exportBase, err := c.rawExportBase.toDirective(); err != nil {
		return nil, err
	} else {
		artifactExport.ExportBase = exportBase
	}

	artifactExport.raw = c

	if err := artifactExport.validate(); err != nil {
		return nil, err
	}

	return artifactExport, nil
}
