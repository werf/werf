package config

import "github.com/flant/dapp/pkg/config/ruby_marshal_config"

type DimgArtifact struct {
	*DimgBase
	Shell *ShellArtifact
}

func (c *DimgArtifact) DimgTree() (tree []DimgInterface) {
	if c.FromDimg != nil {
		tree = append(tree, c.FromDimg.DimgTree()...)
	}
	if c.FromDimgArtifact != nil {
		tree = append(tree, c.FromDimgArtifact.DimgTree()...)
	}

	for _, importElm := range c.Import {
		tree = append(tree, importElm.artifactDimg.DimgTree()...)
	}

	tree = append(tree, c)

	return
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

func (c *DimgArtifact) lastLayerOrSelf() DimgInterface {
	if c.FromDimg != nil {
		return c.FromDimg.lastLayerOrSelf()
	}
	if c.FromDimgArtifact != nil {
		return c.FromDimgArtifact.lastLayerOrSelf()
	}
	return c
}

func (c *DimgArtifact) validate() error {
	if !oneOrNone([]bool{c.Shell != nil, c.Ansible != nil}) {
		return newDetailedConfigError("Cannot use shell and ansible builders at the same time!", nil, c.DimgBase.raw.doc)
	}

	return nil
}

func (c *DimgArtifact) toRuby() ruby_marshal_config.DimgArtifact {
	return *c.toRubyPointer()
}

func (c *DimgArtifact) toRubyPointer() *ruby_marshal_config.DimgArtifact {
	rubyArtifactDimg := &ruby_marshal_config.DimgArtifact{}
	rubyArtifactDimg.DimgBase = c.DimgBase.toRuby()
	rubyArtifactDimg.Name = c.Name
	rubyArtifactDimg.Docker.From = c.From
	rubyArtifactDimg.Docker.FromCacheVersion = c.FromCacheVersion

	if c.Shell != nil {
		rubyArtifactDimg.Shell = c.Shell.toRuby()
	}

	return rubyArtifactDimg
}
