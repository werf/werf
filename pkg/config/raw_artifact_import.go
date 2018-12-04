package config

type rawArtifactImport struct {
	ArtifactName string `yaml:"artifact,omitempty"`
	Before       string `yaml:"before,omitempty"`
	After        string `yaml:"after,omitempty"`

	rawArtifactExport `yaml:",inline"`
	rawDimg           *rawDimg `yaml:"-"` // parent

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *rawArtifactImport) configSection() interface{} {
	return c
}

func (c *rawArtifactImport) doc() *doc {
	return c.rawDimg.doc
}

func (c *rawArtifactImport) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if parent, ok := parentStack.Peek().(*rawDimg); ok {
		c.rawDimg = parent
	}

	parentStack.Push(c)
	type plain rawArtifactImport
	err := unmarshal((*plain)(c))
	parentStack.Pop()
	if err != nil {
		return err
	}

	c.rawArtifactExport.inlinedIntoRaw(c)

	if err := checkOverflow(c.UnsupportedAttributes, c, c.rawDimg.doc); err != nil {
		return err
	}

	if c.rawArtifactExport.rawExportBase.To == "" {
		c.rawArtifactExport.rawExportBase.To = c.rawArtifactExport.rawExportBase.Add
	}

	return nil
}

func (c *rawArtifactImport) toDirective() (artifactImport *ArtifactImport, err error) {
	artifactImport = &ArtifactImport{}

	if artifactExport, err := c.rawArtifactExport.toDirective(); err != nil {
		return nil, err
	} else {
		artifactImport.ArtifactExport = artifactExport
	}

	artifactImport.ArtifactName = c.ArtifactName
	artifactImport.Before = c.Before
	artifactImport.After = c.After

	artifactImport.raw = c

	if err = c.validateDirective(artifactImport); err != nil {
		return nil, err
	}

	return artifactImport, nil
}

func (c *rawArtifactImport) validateDirective(artifactImport *ArtifactImport) (err error) {
	if err = artifactImport.validate(); err != nil {
		return err
	}

	return nil
}
