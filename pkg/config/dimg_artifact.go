package config

import "github.com/flant/dapp/pkg/config/ruby_marshal_config"

type DimgArtifact struct {
	*DimgBase
	Shell *ShellArtifact
}

func (c *DimgArtifact) RelatedDimgs() (relatedDimgs []DimgInterface) {
	relatedDimgs = append(relatedDimgs, c)
	if c.FromDimg != nil {
		relatedDimgs = append(relatedDimgs, c.FromDimg.RelatedDimgs()...)
	}
	if c.FromDimgArtifact != nil {
		relatedDimgs = append(relatedDimgs, c.FromDimgArtifact.RelatedDimgs()...)
	}
	return
}

func (c *DimgArtifact) LastLayerOrSelf() DimgInterface {
	if c.FromDimg != nil {
		return c.FromDimg.LastLayerOrSelf()
	}
	if c.FromDimgArtifact != nil {
		return c.FromDimgArtifact.LastLayerOrSelf()
	}
	return c
}

func (c *DimgArtifact) Validate() error {
	if !OneOrNone([]bool{c.Shell != nil, c.Ansible != nil}) {
		return NewDetailedConfigError("Cannot use shell and ansible builders at the same time!", nil, c.DimgBase.Raw.Doc)
	}

	return nil
}

func (c *DimgArtifact) ToRuby() ruby_marshal_config.DimgArtifact {
	return *c.ToRubyPointer()
}

func (c *DimgArtifact) ToRubyPointer() *ruby_marshal_config.DimgArtifact {
	rubyArtifactDimg := &ruby_marshal_config.DimgArtifact{}
	rubyArtifactDimg.DimgBase = c.DimgBase.ToRuby()
	rubyArtifactDimg.Name = c.Name
	rubyArtifactDimg.Docker.From = c.From
	rubyArtifactDimg.Docker.FromCacheVersion = c.FromCacheVersion

	if c.Shell != nil {
		rubyArtifactDimg.Shell = c.Shell.ToRuby()
	}

	return rubyArtifactDimg
}
