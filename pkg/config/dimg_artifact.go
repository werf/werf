package config

import (
	"fmt"
	"github.com/flant/dapp/pkg/config/ruby_marshal_config"
)

type DimgArtifact struct {
	*DimgBase
	Shell *ShellArtifact
}

func (c *DimgArtifact) Validate() error {
	if c.Chef != nil && c.Shell != nil {
		return fmt.Errorf("Cannot use shell and chef builders at the same time!\n\n%s", DumpConfigDoc(c.DimgBase.Raw.Doc))
	}

	return nil
}

func (c *DimgArtifact) ToRuby() ruby_marshal_config.DimgArtifact {
	rubyArtifactDimg := ruby_marshal_config.DimgArtifact{}
	rubyArtifactDimg.DimgBase = c.DimgBase.ToRuby()
	rubyArtifactDimg.Name = c.Name
	rubyArtifactDimg.Docker.From = c.From

	if c.Shell != nil {
		rubyArtifactDimg.Shell = c.Shell.ToRuby()
	}

	return rubyArtifactDimg
}
