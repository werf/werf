package config

type rawMetaDeployTemplates struct {
	HelmRelease     *string `yaml:"helmRelease,omitempty"`
	HelmReleaseSlug *bool   `yaml:"helmReleaseSlug,omitempty"`
	Namespace       *string `yaml:"namespace,omitempty"`
	NamespaceSlug   *bool   `yaml:"namespaceSlug,omitempty"`

	rawMeta *rawMeta

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *rawMetaDeployTemplates) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if parent, ok := parentStack.Peek().(*rawMeta); ok {
		c.rawMeta = parent
	}

	parentStack.Push(c)
	type plain rawMetaDeployTemplates
	err := unmarshal((*plain)(c))
	parentStack.Pop()
	if err != nil {
		return err
	}

	if err := checkOverflow(c.UnsupportedAttributes, nil, c.rawMeta.doc); err != nil {
		return err
	}

	if c.HelmRelease != nil && *c.HelmRelease == "" {
		return newDetailedConfigError("helmRelease field cannot be empty!", nil, c.rawMeta.doc)
	}

	if c.Namespace != nil && *c.Namespace == "" {
		return newDetailedConfigError("namespace field cannot be empty!", nil, c.rawMeta.doc)
	}

	return nil
}

func (c *rawMetaDeployTemplates) toDeployTemplates() MetaDeployTemplates {
	deployTemplates := MetaDeployTemplates{}
	deployTemplates.HelmRelease = c.HelmRelease
	deployTemplates.HelmReleaseSlug = c.HelmReleaseSlug
	deployTemplates.Namespace = c.Namespace
	deployTemplates.NamespaceSlug = c.NamespaceSlug
	return deployTemplates
}
