package config

type RawArtifactImport struct {
	RawArtifactExportBase `yaml:",inline"`
	ArtifactName          string `yaml:"artifact,omitempty"`
	Before                string `yaml:"before,omitempty"`
	After                 string `yaml:"after,omitempty"`

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *RawArtifactImport) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain RawArtifactImport
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	if err := CheckOverflow(c.UnsupportedAttributes, c); err != nil {
		return err
	}

	return nil
}

func (c *RawArtifactImport) ToDirective() (artifactImport *ArtifactImport, err error) {
	artifactImport = &ArtifactImport{}

	if exportBase, err := c.RawExportBase.ToDirective(); err != nil {
		return nil, err
	} else {
		artifactImport.ExportBase = exportBase
	}

	artifactImport.ArtifactName = c.ArtifactName
	artifactImport.Before = c.Before
	artifactImport.After = c.After

	if err = artifactImport.Validate(); err != nil {
		return nil, err
	}

	return artifactImport, nil
}
