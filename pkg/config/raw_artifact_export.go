package config

type RawArtifactExport struct {
	RawExportBase `yaml:",inline"`
}

func (c *RawArtifactExport) ToDirective() (artifactExport *ArtifactExport, err error) {
	artifactExport = &ArtifactExport{}

	if exportBase, err := c.RawExportBase.ToDirective(); err != nil {
		return nil, err
	} else {
		artifactExport.ExportBase = exportBase
	}

	artifactExport.Raw = c

	if err := c.ValidateDirective(artifactExport); err != nil {
		return nil, err
	}

	return artifactExport, nil
}

func (c *RawArtifactExport) ValidateDirective(artifactExport *ArtifactExport) error {
	if err := artifactExport.Validate(); err != nil {
		return err
	}

	return nil
}
