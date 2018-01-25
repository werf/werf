package config

import (
	"fmt"
)

type DimgBase struct {
	Dimg     interface{}       `yaml:"dimg,omitempty"`
	Artifact string            `yaml:"artifact,omitempty"`
	From     string            `yaml:"from,omitempty"`
	Git      []*Git            `yaml:"git,omitempty"`
	Shell    *ShellBase        `yaml:"shell,omitempty"`
	Chef     *Chef             `yaml:"chef,omitempty"`
	Mount    []*Mount          `yaml:"mount,omitempty"`
	Docker   *Docker           `yaml:"docker,omitempty"`
	Import   []*ArtifactImport `yaml:"import,omitempty"`

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *DimgBase) Type() string {
	_, typeDimg := c.Dimg.(string)
	_, typeDimgArray := c.Dimg.([]string)

	if typeDimg {
		return "dimg"
	} else if typeDimgArray {
		return "dimgArray"
	} else {
		return "artifact"
	}
}

func (c *DimgBase) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain DimgBase
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

func (c *DimgBase) Validate() error {
	if err := c.ValidateRequiredFields(); err != nil {
		return err
	}

	if err := c.ValidateType(); err != nil {
		return err
	}
	return nil
}

func (c *DimgBase) ValidateRequiredFields() error {
	if c.From != "" {
		return fmt.Errorf("from не может быть пустым!") // FIXME
	}
	return nil
}

func (c *DimgBase) ValidateType() error {
	typeArtifact := c.Artifact != ""
	_, typeDimg := c.Dimg.(string)
	_, typeDimgArray := c.Dimg.([]string)

	if typeArtifact && (typeDimg || typeDimgArray) {
		return fmt.Errorf("Conflict between `dimg` and `artifact` directives!") // FIXME
	} else if !(typeArtifact || typeDimg || typeDimgArray) {
		return fmt.Errorf("dimg не имеет связи ни с артефактом ни с dimg") // FIXME
	}
	return nil
}
