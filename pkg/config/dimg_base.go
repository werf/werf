package config

import (
	"fmt"
)

type DimgInterface interface{}

type DimgBase struct {
	DimgInterface

	Name             string
	From             string
	FromDimg         *Dimg
	FromDimgArtifact *DimgArtifact
	FromCacheVersion string
	Git              *GitManager
	Shell            *Shell
	Ansible          *Ansible
	Mount            []*Mount
	Import           []*ArtifactImport

	raw *rawDimg
}

func (c *DimgBase) associateFrom(dimgs []*Dimg, artifacts []*DimgArtifact) error {
	if c.FromDimg != nil || c.FromDimgArtifact != nil { // asLayers
		return nil
	}

	if c.raw.FromDimg != "" {
		fromDimgName := c.raw.FromDimg

		if fromDimgName == c.Name {
			return newDetailedConfigError(fmt.Sprintf("cannot use own dimg name as `fromDimg` directive value!"), nil, c.raw.doc)
		}

		if dimg := dimgByName(dimgs, fromDimgName); dimg != nil {
			c.FromDimg = dimg
		} else {
			return newDetailedConfigError(fmt.Sprintf("no such dimg `%s`!", fromDimgName), c.raw, c.raw.doc)
		}
	} else if c.raw.FromDimgArtifact != "" {
		fromDimgArtifactName := c.raw.FromDimgArtifact

		if fromDimgArtifactName == c.Name {
			return newDetailedConfigError(fmt.Sprintf("cannot use own dimg name as `fromDimgArtifact` directive value!"), nil, c.raw.doc)
		}

		if dimgArtifact := dimgArtifactByName(artifacts, fromDimgArtifactName); dimgArtifact != nil {
			c.FromDimgArtifact = dimgArtifact
		} else {
			return newDetailedConfigError(fmt.Sprintf("no such dimg artifact `%s`!", fromDimgArtifactName), c.raw, c.raw.doc)
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
