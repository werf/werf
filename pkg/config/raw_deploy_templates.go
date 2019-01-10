package config

type rawDeployTemplates struct {
	HelmRelease         string `yaml:"helm_release,omitempty"`
	KubernetesNamespace string `yaml:"kubernetes_namespace,omitempty"`

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

	return nil
}

func (c *rawDeployTemplates) toDeployTemplates() DeployTemplates {
	return DeployTemplates{
		KubernetesNamespace: c.KubernetesNamespace,
		HelmRelease:         c.HelmRelease,
	}
}
