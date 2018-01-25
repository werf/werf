package config

import "fmt"

type ArtifactImport struct {
	ExportBase   `yaml:",inline"`
	ArtifactName string `yaml:"artifact,omitempty"`
	Before       string `yaml:"before,omitempty"`
	After        string `yaml:"after,omitempty"`

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *ArtifactImport) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain ArtifactImport
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

func (c *ArtifactImport) Validate() error {
	if err := c.ExportBase.Validate(); err != nil {
		return err
	}

	if err := c.ValidateRequiredFields(); err != nil {
		return err
	}
	return nil
}

func (c *ArtifactImport) ValidateRequiredFields() error {
	if c.ArtifactName == "" {
		return fmt.Errorf("имя артефакта обязательно") // FIXME
	} else if c.Before == "" && c.After == "" {
		return fmt.Errorf("артефакт должен иметь связь!") // FIXME
	} else if c.Before != "" && checkRelation(c.Before) {
		return fmt.Errorf("артефакт имеет некорректную связь!") // FIXME
	} else if c.After != "" && checkRelation(c.After) {
		return fmt.Errorf("артефакт имеет некорректную связь!") // FIXME
	}
	return nil
}

func checkRelation(relation string) bool {
	return relation == "install" || relation == "setup"
}
