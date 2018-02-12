package config

import (
	"fmt"
	"github.com/flant/dapp/pkg/config/ruby_marshal_config"
)

type DimgBase struct {
	Name    string
	From    string
	Bulder  string
	Git     *GitManager
	Ansible *Ansible
	Chef    *Chef
	Mount   []*Mount
	Import  []*ArtifactImport

	Raw *RawDimg
}

func (c *DimgBase) Validate() error {
	if c.From == "" {
		return fmt.Errorf("`from: DOCKER_IMAGE` required!\n\n%s", DumpConfigDoc(c.Raw.Doc))
	}

	// TODO: валидацию формата `From`
	// TODO: валидация формата `Name`

	return nil
}

func (c *DimgBase) ToRuby() ruby_marshal_config.DimgBase {
	rubyDimg := ruby_marshal_config.DimgBase{}
	rubyDimg.Name = c.Name
	rubyDimg.Builder = ruby_marshal_config.Symbol(c.Bulder)

	if c.Chef != nil {
		rubyDimg.Chef = c.Chef.ToRuby()
	}

	if c.Ansible != nil {
		rubyDimg.Ansible = c.Ansible.ToRuby()
	}

	if c.Git != nil {
		rubyDimg.GitArtifact = c.Git.ToRuby()
	}

	for _, mount := range c.Mount {
		rubyDimg.Mount = append(rubyDimg.Mount, mount.ToRuby())
	}

	for _, importArtifact := range c.Import {
		artifactGroup := ruby_marshal_config.ArtifactGroup{}
		artifactGroup.Export = append(artifactGroup.Export, importArtifact.ToRuby())
		rubyDimg.ArtifactGroup = append(rubyDimg.ArtifactGroup, artifactGroup)
	}

	return rubyDimg
}
