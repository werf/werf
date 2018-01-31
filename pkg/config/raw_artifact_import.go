package config

type RawArtifactImport struct {
	RawArtifactExport `yaml:",inline"`
	ArtifactName      string `yaml:"artifact,omitempty"`
	Before            string `yaml:"before,omitempty"`
	After             string `yaml:"after,omitempty"`

	RawDimg *RawDimg `yaml:"-"` // parent

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *RawArtifactImport) ConfigSection() interface{} {
	return c
}

func (c *RawArtifactImport) Doc() *Doc {
	return c.RawDimg.Doc
}

func (c *RawArtifactImport) UnmarshalYAML(unmarshal func(interface{}) error) error {
	c.RawArtifactExport.RawExportBase = NewRawExportBase()
	if parent, ok := ParentStack.Peek().(*RawDimg); ok {
		c.RawDimg = parent
	}

	ParentStack.Push(c)
	type plain RawArtifactImport
	err := unmarshal((*plain)(c))
	ParentStack.Pop()
	if err != nil {
		return err
	}

	c.RawArtifactExport.InlinedIntoRaw(c)

	if err := CheckOverflow(c.UnsupportedAttributes, c, c.RawDimg.Doc); err != nil {
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
