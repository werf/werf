package config

import (
	"fmt"
)

type RawShell struct {
	BeforeInstall interface{} `yaml:"beforeInstall,omitempty"`
	Install       interface{} `yaml:"install,omitempty"`
	BeforeSetup   interface{} `yaml:"beforeSetup,omitempty"`
	Setup         interface{} `yaml:"setup,omitempty"`
	BuildArtifact interface{} `yaml:"buildArtifact,omitempty"`

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *RawShell) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain RawShell
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	if err := CheckOverflow(c.UnsupportedAttributes, c); err != nil {
		return err
	}

	return nil
}

func (c *RawShell) ToDirective() (shellDimg *ShellDimg, err error) {
	shellDimg = &ShellDimg{}
	shellDimg.ShellBase = &ShellBase{}

	if beforeInstall, err := InterfaceToStringArray(c.BeforeInstall); err != nil {
		return nil, err
	} else {
		shellDimg.ShellBase.BeforeInstall = beforeInstall
	}

	if install, err := InterfaceToStringArray(c.Install); err != nil {
		return nil, err
	} else {
		shellDimg.ShellBase.Install = install
	}

	if beforeSetup, err := InterfaceToStringArray(c.BeforeSetup); err != nil {
		return nil, err
	} else {
		shellDimg.ShellBase.BeforeSetup = beforeSetup
	}

	if setup, err := InterfaceToStringArray(c.Setup); err != nil {
		return nil, err
	} else {
		shellDimg.ShellBase.Setup = setup
	}

	shellDimg.ShellBase.Raw = c

	if err := c.ValidateDirective(shellDimg); err != nil {
		return nil, err
	}

	return shellDimg, nil
}

func (c *RawShell) ValidateDirective(shellDimg *ShellDimg) error {
	if c.BuildArtifact != nil {
		return fmt.Errorf("директива buildArtifact не может быть объявлена для dimg-а!") // FIXME
	}

	if err := shellDimg.Validate(); err != nil {
		return err
	}

	return nil
}

func (c *RawShell) ToArtifactDirective() (shellArtifact *ShellArtifact, err error) {
	shellArtifact = &ShellArtifact{}

	if shellDimg, err := c.ToDirective(); err != nil {
		return nil, err
	} else {
		shellArtifact.ShellDimg = shellDimg
	}

	if err := c.ValidateArtifactDirective(shellArtifact); err != nil {
		return nil, err
	}

	return shellArtifact, nil
}

func (c *RawShell) ValidateArtifactDirective(shellArtifact *ShellArtifact) error {
	if err := shellArtifact.Validate(); err != nil {
		return err
	}

	return nil
}
