package config

import (
	"fmt"
)

type Git struct {
	ExportBase        `yaml:",inline"`
	As                string             `yaml:"as,omitempty"`
	Url               string             `yaml:"url,omitempty"`
	Branch            string             `yaml:"branch,omitempty"`
	Commit            string             `yaml:"commit,omitempty"`
	StageDependencies *StageDependencies `yaml:"stageDependencies,omitempty"`

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *Git) Type() string {
	if c.Url != "" {
		return "remote"
	}
	return "local"
}

func (c *Git) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Git
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	if err := CheckOverflow(c.UnsupportedAttributes, c); err != nil {
		return err
	}

	return nil
}

func (c *Git) Validate() error {
	if err := c.ExportBase.Validate(); err != nil {
		return err
	}
	return nil
}

func (c *Git) ToRuby() interface{} {
	return nil // TODO
}

type StageDependencies struct {
	Install       interface{} `yaml:"install,omitempty"`
	Setup         interface{} `yaml:"setup,omitempty"`
	BeforeSetup   interface{} `yaml:"beforeSetup,omitempty"`
	BuildArtifact interface{} `yaml:"buildArtifact,omitempty"`

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *StageDependencies) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain StageDependencies
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	if err := CheckOverflow(c.UnsupportedAttributes, c); err != nil {
		return err
	}

	return nil
}

type ExportBase struct {
	Add          string      `yaml:"add,omitempty"`
	To           string      `yaml:"to,omitempty"`
	IncludePaths interface{} `yaml:"includePaths,omitempty"`
	ExcludePaths interface{} `yaml:"excludePaths,omitempty"`
	Owner        string      `yaml:"owner,omitempty"`
	Group        string      `yaml:"group,omitempty"`
}

func (c *ExportBase) Validate() error {
	if c.To == "" {
		return fmt.Errorf("to не может быть пустым!") // FIXME
	}
	return nil
}
