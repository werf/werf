package config

type rawMeta struct {
	Project         *string            `yaml:"project,omitempty"`
	DeployTemplates rawDeployTemplates `yaml:"deploy,omitempty"`

	doc *doc `yaml:"-"` // parent

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *rawMeta) UnmarshalYAML(unmarshal func(interface{}) error) error {
	parentStack.Push(c)
	type plain rawMeta
	err := unmarshal((*plain)(c))
	parentStack.Pop()
	if err != nil {
		return err
	}

	if err := checkOverflow(c.UnsupportedAttributes, nil, c.doc); err != nil {
		return err
	}

	if c.Project != nil && *c.Project == "" {
		return newDetailedConfigError("project field cannot be empty!", nil, c.doc)
	}

	return nil
}

func (c *rawMeta) toMeta() *Meta {
	meta := &Meta{}
	if c.Project != nil {
		meta.Project = *c.Project
	}

	meta.DeployTemplates = c.DeployTemplates.toDeployTemplates()

	return meta
}
