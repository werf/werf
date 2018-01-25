package config

import (
	"fmt"
	"github.com/flant/dapp/pkg/config/ruby_marshal_config"
)

type ShellBase struct {
	BeforeInstall interface{} `yaml:"beforeInstall,omitempty"`
	Install       interface{} `yaml:"install,omitempty"`
	BeforeSetup   interface{} `yaml:"beforeSetup,omitempty"`
	Setup         interface{} `yaml:"setup,omitempty"`
	BuildArtifact interface{} `yaml:"buildArtifact,omitempty"`

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *ShellBase) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain ShellBase
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	if err := CheckOverflow(c.UnsupportedAttributes, c); err != nil {
		return err
	}

	return nil
}

type ShellDimg struct{ ShellBase }

func (c *ShellDimg) ValidateDirectives() error {
	if c.BuildArtifact != nil {
		return fmt.Errorf("директива build artifact не может быть объявлена в dimg-е") // FIXME
	}
	return nil
}

func (c *ShellDimg) ToRuby() ruby_marshal_config.ShellDimg {
	var shellDimg ruby_marshal_config.ShellDimg

	if c.BeforeInstall != nil {
		shellDimg.BeforeInstall.Run = runCommands(c.BeforeInstall)
	}
	if c.Install != nil {
		shellDimg.Install.Run = runCommands(c.Install)
	}
	if c.BeforeSetup != nil {
		shellDimg.BeforeSetup.Run = runCommands(c.BeforeSetup)
	}
	if c.Setup != nil {
		shellDimg.Setup.Run = runCommands(c.Setup)
	}

	return shellDimg
}

type ShellArtifact struct{ ShellDimg }

func (c *ShellArtifact) ValidateDirectives() error {
	return nil
}

func (c *ShellArtifact) ToRuby() ruby_marshal_config.ShellArtifact {
	shellDimg := c.ShellDimg.ToRuby()
	shellArtifact := ruby_marshal_config.ShellArtifact{ShellDimg: shellDimg}

	if c.BuildArtifact != nil {
		shellArtifact.BuildArtifact.Run = runCommands(c.BuildArtifact)
	}
	return shellArtifact
}

func runCommands(directive interface{}) []string {
	if val, ok := directive.(string); ok {
		return []string{val}
	} else if val, ok := directive.([]string); ok {
		return val
	}
	return nil
}
