package config

import (
	"github.com/flant/dapp/pkg/config/directive"
)

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

	return nil
}

func (c *Mount) ToDirective() (mount *config.Mount, err error) {
	mount = &config.Mount{}
	mount.From = c.From
	mount.To = c.To

	if err := mount.Validate(); err != nil {
		return nil, err
	}

	return mount, nil
}
