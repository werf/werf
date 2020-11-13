package config

type rawMetaDeploy struct {
	HelmChartDir    *string `yaml:"helmChartDir,omitempty"`
	HelmRelease     *string `yaml:"helmRelease,omitempty"`
	HelmReleaseSlug *bool   `yaml:"helmReleaseSlug,omitempty"`
	Namespace       *string `yaml:"namespace,omitempty"`
	NamespaceSlug   *bool   `yaml:"namespaceSlug,omitempty"`

	rawMeta *rawMeta

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *rawMetaDeploy) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if parent, ok := parentStack.Peek().(*rawMeta); ok {
		c.rawMeta = parent
	}

	parentStack.Push(c)
	type plain rawMetaDeploy
	err := unmarshal((*plain)(c))
	parentStack.Pop()
	if err != nil {
		return err
	}

	if err := checkOverflow(c.UnsupportedAttributes, nil, c.rawMeta.doc); err != nil {
		return err
	}

	if c.HelmChartDir != nil && *c.HelmChartDir == "" {
		return newDetailedConfigError("helmChartDir field cannot be empty!", nil, c.rawMeta.doc)
	}

	if c.HelmRelease != nil && *c.HelmRelease == "" {
		return newDetailedConfigError("helmRelease field cannot be empty!", nil, c.rawMeta.doc)
	}

	if c.Namespace != nil && *c.Namespace == "" {
		return newDetailedConfigError("namespace field cannot be empty!", nil, c.rawMeta.doc)
	}

	return nil
}

func (c *rawMetaDeploy) toMetaDeploy() MetaDeploy {
	metaDeploy := MetaDeploy{}
	metaDeploy.HelmChartDir = c.HelmChartDir
	metaDeploy.HelmRelease = c.HelmRelease
	metaDeploy.HelmReleaseSlug = c.HelmReleaseSlug
	metaDeploy.Namespace = c.Namespace
	metaDeploy.NamespaceSlug = c.NamespaceSlug
	return metaDeploy
}
