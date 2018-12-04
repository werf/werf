package config

import "fmt"

type rawMount struct {
	To       string `yaml:"to,omitempty"`
	From     string `yaml:"from,omitempty"`
	FromPath string `yaml:"fromPath,omitempty"`

	rawDimg *rawDimg `yaml:"-"` // parent

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *rawMount) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if parent, ok := parentStack.Peek().(*rawDimg); ok {
		c.rawDimg = parent
	}

	type plain rawMount
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	if err := checkOverflow(c.UnsupportedAttributes, c, c.rawDimg.doc); err != nil {
		return err
	}

	return nil
}

func (c *rawMount) toDirective() (mount *Mount, err error) {
	mount = &Mount{}
	mount.To = c.To
	mount.From = c.FromPath

	if c.From == "" {
		mount.Type = "custom_dir"
	} else {
		mount.Type = c.From
	}

	mount.raw = c

	if err := c.validateDirective(mount); err != nil {
		return nil, err
	}

	return mount, nil
}

func (c *rawMount) validateDirective(mount *Mount) (err error) {
	if c.From != "" && c.FromPath != "" {
		return newDetailedConfigError(fmt.Sprintf("Cannot use `from: %s` and `fromPath: %s` at the same time for mount!", c.From, c.FromPath), c, c.rawDimg.doc)
	}

	if err := mount.validate(); err != nil {
		return err
	}

	return nil
}
