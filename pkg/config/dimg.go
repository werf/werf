package config

import "github.com/flant/dapp/pkg/config/ruby_marshal_config"

type Dimg struct {
	*DimgBase
	Shell  *ShellDimg
	Docker *Docker
}

func (c *Dimg) RelatedDimgs() (relatedDimgs []DimgInterface) {
	relatedDimgs = append(relatedDimgs, c)
	if c.FromDimg != nil {
		relatedDimgs = append(relatedDimgs, c.FromDimg.RelatedDimgs()...)
	}
	if c.FromDimgArtifact != nil {
		relatedDimgs = append(relatedDimgs, c.FromDimgArtifact.RelatedDimgs()...)
	}
	return
}

func (c *Dimg) lastLayerOrSelf() DimgInterface {
	if c.FromDimg != nil {
		return c.FromDimg.lastLayerOrSelf()
	}
	if c.FromDimgArtifact != nil {
		return c.FromDimgArtifact.lastLayerOrSelf()
	}
	return c
}

func (c *Dimg) validate() error {
	if !oneOrNone([]bool{c.Shell != nil, c.Ansible != nil}) {
		return newDetailedConfigError("Cannot use shell and ansible builders at the same time!", nil, c.DimgBase.raw.doc)
	}

	return nil
}

func (c *Dimg) toRuby() ruby_marshal_config.Dimg {
	return *c.toRubyPointer()
}

func (c *Dimg) toRubyPointer() *ruby_marshal_config.Dimg {
	rubyDimg := &ruby_marshal_config.Dimg{}
	rubyDimg.DimgBase = c.DimgBase.toRuby()

	if c.Shell != nil {
		rubyDimg.Shell = c.Shell.toRuby()
	}

	if c.Docker != nil {
		rubyDimg.Docker = c.Docker.toRuby()
	}
	rubyDimg.Docker.From = c.From
	rubyDimg.Docker.FromCacheVersion = c.FromCacheVersion

	return rubyDimg
}
