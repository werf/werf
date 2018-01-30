package config

type RawMount struct {
	From string `yaml:"from,omitempty"`
	To   string `yaml:"to,omitempty"`

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *RawMount) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain RawMount
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	if err := CheckOverflow(c.UnsupportedAttributes, c); err != nil {
		return err
	}

	return nil
}

func (c *RawMount) ToDirective() (mount *Mount, err error) {
	mount = &Mount{}
	mount.From = c.From
	mount.To = c.To

	if err := mount.Validate(); err != nil {
		return nil, err
	}

	return mount, nil
}
