package config

type RawChef struct {
	Cookbook   string                      `yaml:"cookbook,omitempty"`
	Recipe     interface{}                 `yaml:"recipe,omitempty"`
	Attributes map[interface{}]interface{} `yaml:"attributes,omitempty"`

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *RawChef) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain RawChef
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	if err := CheckOverflow(c.UnsupportedAttributes, c); err != nil {
		return err
	}

	return nil
}

func (c *RawChef) ToDirective() (chef *Chef, err error) {
	chef = &Chef{}
	chef.Cookbook = c.Cookbook

	if recipe, err := InterfaceToStringArray(c.Recipe); err != nil {
		return nil, err
	} else {
		chef.Recipe = recipe
	}

	chef.Attributes = c.Attributes
	return chef, nil
}
