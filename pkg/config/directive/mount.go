package config

import "fmt"

type Mount struct {
	From string `yaml:"from,omitempty"`
	To   string `yaml:"to,omitempty"`

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *Mount) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Mount
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	if err := CheckOverflow(c.UnsupportedAttributes, c); err != nil {
		return err
	}

	if err := c.Validate(); err != nil {
		return err
	}

	return nil
}

func (c *Mount) Validate() error {
	if err := c.ValidateRequiredFields(); err != nil {
		return err
	}
	return nil
}

func (c *Mount) ValidateRequiredFields() error {
	if c.From == "" {
		return fmt.Errorf("from не может быть пустым!") // FIXME
	} else if c.To == "" {
		return fmt.Errorf("to не может быть пустым!") // FIXME
	}
	return nil
}
