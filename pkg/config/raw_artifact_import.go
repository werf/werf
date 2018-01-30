package config

type RawArtifactImport struct {
	RawArtifactExport `yaml:",inline"`
	ArtifactName      string `yaml:"artifact,omitempty"`
	Before            string `yaml:"before,omitempty"`
	After             string `yaml:"after,omitempty"`

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

	if artifactExport, err := c.RawArtifactExport.ToDirective(); err != nil {
		return nil, err
	} else {
		artifactImport.ArtifactExport = artifactExport
	}

	artifactImport.ArtifactName = c.ArtifactName
	artifactImport.Before = c.Before
	artifactImport.After = c.After

	artifactImport.Raw = c

	if err = c.ValidateDirective(artifactImport); err != nil {
		return nil, err
	}

	return artifactImport, nil
}

func (c *RawArtifactImport) ValidateDirective(artifactImport *ArtifactImport) (err error) {
	if err = artifactImport.Validate(); err != nil {
		return err
	}

	return nil
}
