package config

import (
	"fmt"
	"github.com/flant/dapp/pkg/config/ruby_marshal_config"
)

type DimgInterface interface{}

type DimgBase struct {
	DimgInterface

	Name             string
	From             string
	FromCacheVersion string
	FromDimg         *Dimg
	FromDimgArtifact *DimgArtifact
	Bulder           string
	Git              *GitManager
	Ansible          *Ansible
	Mount            []*Mount
	Import           []*ArtifactImport

	Raw *RawDimg
}

func (c *DimgBase) AssociateFrom(dimgs []*Dimg, artifacts []*DimgArtifact) error {
	if c.FromDimg != nil || c.FromDimgArtifact != nil { // asLayers
		return nil
	}

	if c.Raw.FromDimg != "" {
		fromDimgName := c.Raw.FromDimg

		if fromDimgName == c.Name {
			return NewDetailedConfigError(fmt.Sprintf("Cannot use own dimg name as `fromDimg` directive value!"), nil, c.Raw.Doc)
		}

		if dimg := dimgByName(dimgs, fromDimgName); dimg != nil {
			c.FromDimg = dimg
		} else {
			return NewDetailedConfigError(fmt.Sprintf("No such dimg `%s`!", fromDimgName), c.Raw, c.Raw.Doc)
		}
	} else if c.Raw.FromDimgArtifact != "" {
		fromDimgArtifactName := c.Raw.FromDimgArtifact

		if fromDimgArtifactName == c.Name {
			return NewDetailedConfigError(fmt.Sprintf("Cannot use own dimg name as `fromDimgArtifact` directive value!"), nil, c.Raw.Doc)
		}

		if dimgArtifact := dimgArtifactByName(artifacts, fromDimgArtifactName); dimgArtifact != nil {
			c.FromDimgArtifact = dimgArtifact
		} else {
			return NewDetailedConfigError(fmt.Sprintf("No such dimg artifact `%s`!", fromDimgArtifactName), c.Raw, c.Raw.Doc)
		}
	}

	return nil
}

func dimgByName(dimgs []*Dimg, name string) *Dimg {
	for _, dimg := range dimgs {
		if dimg.Name == name {
			return dimg
		}
	}
	return nil
}

func dimgArtifactByName(dimgs []*DimgArtifact, name string) *DimgArtifact {
	for _, dimg := range dimgs {
		if dimg.Name == name {
			return dimg
		}
	}
	return nil
}

func (c *DimgBase) Validate() error {
	if c.From == "" && c.Raw.FromDimg == "" && c.Raw.FromDimgArtifact == "" && c.FromDimg == nil && c.FromDimgArtifact == nil {
		return NewDetailedConfigError("`from: DOCKER_IMAGE`, `fromDimg: DIMG_NAME`, `fromDimgArtifact: ARTIFACT_DIMG_NAME` required!", nil, c.Raw.Doc)
	}

	if !OneOrNone([]bool{c.From != "", c.Raw.FromDimg != "", c.Raw.FromDimgArtifact != ""}) {
		return NewDetailedConfigError("`conflict between `from`, `fromDimg` and `fromDimgArtifact` directives!", nil, c.Raw.Doc)
	}

	// TODO: валидацию формата `From`
	// TODO: валидация формата `Name`

	return nil
}

func (c *DimgBase) ToRuby() ruby_marshal_config.DimgBase {
	rubyDimg := ruby_marshal_config.DimgBase{}
	rubyDimg.Name = c.Name
	rubyDimg.Builder = ruby_marshal_config.Symbol(c.Bulder)

	if c.FromDimg != nil {
		rubyDimg.FromDimg = c.FromDimg.ToRubyPointer()
	}

	if c.FromDimgArtifact != nil {
		rubyDimg.FromDimgArtifact = c.FromDimgArtifact.ToRubyPointer()
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
