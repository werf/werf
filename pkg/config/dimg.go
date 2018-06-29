package config

import "github.com/flant/dapp/pkg/config/ruby_marshal_config"

type Dimg struct {
	*DimgBase
	Shell  *ShellDimg
	Docker *Docker
}

func (c *Dimg) RelatedDimgs() (relatedDimgs []interface{}) {
	relatedDimgs = append(relatedDimgs, c)
	if c.FromDimg != nil {
		relatedDimgs = append(relatedDimgs, c.FromDimg.RelatedDimgs()...)
	}
	if c.FromDimgArtifact != nil {
		relatedDimgs = append(relatedDimgs, c.FromDimgArtifact.RelatedDimgs()...)
	}
	return
}

func (c *Dimg) Validate() error {
	if !OneOrNone([]bool{c.Shell != nil, c.Ansible != nil}) {
		return NewDetailedConfigError("Cannot use shell and ansible builders at the same time!", nil, c.DimgBase.Raw.Doc)
	}

	return nil
}

func (c *Dimg) ToRuby() ruby_marshal_config.Dimg {
	return *c.ToRubyPointer()
}

func (c *Dimg) ToRubyPointer() *ruby_marshal_config.Dimg {
	rubyDimg := &ruby_marshal_config.Dimg{}
	rubyDimg.DimgBase = c.DimgBase.ToRuby()

	if c.Shell != nil {
		rubyDimg.Shell = c.Shell.ToRuby()
	}

	if c.Docker != nil {
		rubyDimg.Docker = c.Docker.ToRuby()
	}
	rubyDimg.Docker.From = c.From
	rubyDimg.Docker.FromCacheVersion = c.FromCacheVersion

	return rubyDimg
}
