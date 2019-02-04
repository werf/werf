package config

type rawDeployTemplates struct {
	HelmRelease     *string `yaml:"helmRelease,omitempty"`
	HelmReleaseSlug *bool   `yaml:"helmReleaseSlug,omitempty"`
	Namespace       *string `yaml:"namespace,omitempty"`
	NamespaceSlug   *bool   `yaml:"namespaceSlug,omitempty"`

	rawMeta *rawMeta

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *rawDeployTemplates) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if parent, ok := parentStack.Peek().(*rawMeta); ok {
		c.rawMeta = parent
	}

	parentStack.Push(c)
	type plain rawDeployTemplates
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

func (c *rawDeployTemplates) toDeployTemplates() DeployTemplates {
	deployTemplates := DeployTemplates{}

	if c.HelmRelease != nil {
		deployTemplates.HelmRelease = *c.HelmRelease
	}

	deployTemplates.HelmReleaseSlug = true
	if c.HelmReleaseSlug != nil {
		deployTemplates.HelmReleaseSlug = *c.HelmReleaseSlug
	}

	if c.Namespace != nil {
		deployTemplates.Namespace = *c.Namespace
	}

	deployTemplates.NamespaceSlug = true
	if c.NamespaceSlug != nil {
		deployTemplates.NamespaceSlug = *c.NamespaceSlug
	}

	return deployTemplates
}
