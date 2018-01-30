package config

import (
	"fmt"
	"github.com/flant/dapp/pkg/config/ruby_marshal_config"
)

type DimgBase struct {
	Name   string
	From   string
	Bulder string
	Git    *GitManager
	Chef   *Chef
	Mount  []*Mount
	Import []*ArtifactImport

	Raw *RawDimg
}

func (c *DimgBase) Validate() error {
	if c.From == "" {
		return fmt.Errorf("`from` required!") // FIXME
	}

	// TODO: валидацию формата `From`
	// TODO: валидация формата `Name`

	return nil
}

func (c *DimgBase) ToRuby() ruby_marshal_config.DimgBase {
	rubyDimg := ruby_marshal_config.DimgBase{}
	rubyDimg.Name = c.Name
	rubyDimg.Builder = c.Bulder

	if c.Chef != nil {
		rubyDimg.Chef = c.Chef.ToRuby()
	}

	if c.Git != nil {
		rubyDimg.GitArtifact = c.Git.ToRuby()
	}

	for _, mount := range c.Mount {
		rubyDimg.Mount = append(rubyDimg.Mount, mount.ToRuby())
	}

	for _, importArtifact := range c.Import {
		rubyDimg.Artifact = append(rubyDimg.Artifact, importArtifact.ToRuby())
	}

	return rubyDimg
}
