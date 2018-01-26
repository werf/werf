package config

import (
	"github.com/flant/dapp/pkg/config/directive"
)

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

	return nil
}

func (c *ArtifactImport) ToDirective() (artifactImport *config.ArtifactImport, err error) {
	artifactImport = &config.ArtifactImport{}

	if exportBase, err := c.ExportBase.ToDirective(); err != nil {
		return nil, err
	} else {
		artifactImport.ExportBase = exportBase
	}

	artifactImport.ArtifactName = c.ArtifactName
	artifactImport.Before = c.Before
	artifactImport.After = c.After

	if err = artifactImport.Validate(); err != nil {
		return nil, err
	}

	return artifactImport, nil
}
