package config

type RawChef struct {
	Cookbook   string                      `yaml:"cookbook,omitempty"`
	Recipe     interface{}                 `yaml:"recipe,omitempty"`
	Attributes map[interface{}]interface{} `yaml:"attributes,omitempty"`

	RawDimg *RawDimg `yaml:"-"` // parent

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *RawChef) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if parent, ok := ParentStack.Peek().(*RawDimg); ok {
		c.RawDimg = parent
	}

	type plain RawChef
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	if err := CheckOverflow(c.UnsupportedAttributes, c, c.RawDimg.Doc); err != nil {
		return err
	}

	return nil
}

func (c *RawChef) ToDirective() (chef *Chef, err error) {
	chef = &Chef{}
	chef.Cookbook = c.Cookbook

	if recipe, err := InterfaceToStringArray(c.Recipe, c, c.RawDimg.Doc); err != nil {
		return nil, err
	} else {
		chef.Recipe = recipe
	}

	chef.Attributes = c.Attributes

	chef.Raw = c

	if err := c.ValidateDirective(chef); err != nil {
		return nil, err
	}

	return chef, nil
}

func (c *RawChef) ValidateDirective(chef *Chef) (err error) {
	if err := chef.Validate(); err != nil {
		return err
	}

	return nil
}
