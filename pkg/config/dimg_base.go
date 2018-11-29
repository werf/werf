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
	FromDimg         *Dimg         // FIXME: reject in golang binary
	FromDimgArtifact *DimgArtifact // FIXME: FromDimgName field only
	FromCacheVersion string
	Git              *GitManager
	Ansible          *Ansible
	Mount            []*Mount
	Import           []*ArtifactImport

	raw *rawDimg

	builder string // FIXME: reject in golang binary
}

func (c *DimgBase) associateFrom(dimgs []*Dimg, artifacts []*DimgArtifact) error {
	if c.FromDimg != nil || c.FromDimgArtifact != nil { // asLayers
		return nil
	}

	if c.raw.FromDimg != "" {
		fromDimgName := c.raw.FromDimg

		if fromDimgName == c.Name {
			return newDetailedConfigError(fmt.Sprintf("Cannot use own dimg name as `fromDimg` directive value!"), nil, c.raw.doc)
		}

		if dimg := dimgByName(dimgs, fromDimgName); dimg != nil {
			c.FromDimg = dimg
		} else {
			return newDetailedConfigError(fmt.Sprintf("No such dimg `%s`!", fromDimgName), c.raw, c.raw.doc)
		}
	} else if c.raw.FromDimgArtifact != "" {
		fromDimgArtifactName := c.raw.FromDimgArtifact

		if fromDimgArtifactName == c.Name {
			return newDetailedConfigError(fmt.Sprintf("Cannot use own dimg name as `fromDimgArtifact` directive value!"), nil, c.raw.doc)
		}

		if dimgArtifact := dimgArtifactByName(artifacts, fromDimgArtifactName); dimgArtifact != nil {
			c.FromDimgArtifact = dimgArtifact
		} else {
			return newDetailedConfigError(fmt.Sprintf("No such dimg artifact `%s`!", fromDimgArtifactName), c.raw, c.raw.doc)
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

func (c *DimgBase) validate() error {
	if c.From == "" && c.raw.FromDimg == "" && c.raw.FromDimgArtifact == "" && c.FromDimg == nil && c.FromDimgArtifact == nil {
		return newDetailedConfigError("`from: DOCKER_IMAGE`, `fromDimg: DIMG_NAME`, `fromDimgArtifact: ARTIFACT_DIMG_NAME` required!", nil, c.raw.doc)
	}

	if !oneOrNone([]bool{c.From != "", c.raw.FromDimg != "", c.raw.FromDimgArtifact != ""}) {
		return newDetailedConfigError("`conflict between `from`, `fromDimg` and `fromDimgArtifact` directives!", nil, c.raw.doc)
	}

	// TODO: валидацию формата `From`
	// TODO: валидация формата `Name`

	return nil
}

func (c *DimgBase) toRuby() ruby_marshal_config.DimgBase {
	rubyDimg := ruby_marshal_config.DimgBase{}
	rubyDimg.Name = c.Name
	rubyDimg.Builder = ruby_marshal_config.Symbol(c.builder)

	if c.FromDimg != nil {
		rubyDimg.FromDimg = c.FromDimg.toRubyPointer()
	}

	if c.FromDimgArtifact != nil {
		rubyDimg.FromDimgArtifact = c.FromDimgArtifact.toRubyPointer()
	}

	if c.Ansible != nil {
		rubyDimg.Ansible = c.Ansible.toRuby()
	}

	if c.Git != nil {
		rubyDimg.GitArtifact = c.Git.toRuby()
	}

	for _, mount := range c.Mount {
		rubyDimg.Mount = append(rubyDimg.Mount, mount.toRuby())
	}

	for _, importArtifact := range c.Import {
		artifactGroup := ruby_marshal_config.ArtifactGroup{}
		artifactGroup.Export = append(artifactGroup.Export, importArtifact.toRuby())
		rubyDimg.ArtifactGroup = append(rubyDimg.ArtifactGroup, artifactGroup)
	}

	return rubyDimg
}
