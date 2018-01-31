package config

import (
	"fmt"
	"github.com/flant/dapp/pkg/config/ruby_marshal_config"
)

type Dimg struct {
	*DimgBase
	Shell  *ShellDimg
	Docker *Docker
}

func (c *Dimg) Validate() error {
	if c.Chef != nil && c.Shell != nil {
		return fmt.Errorf("Cannot use shell and chef builders at the same time!") // FIXME
	}

	return nil
}

func (c *Dimg) ToRuby() ruby_marshal_config.Dimg {
	rubyDimg := ruby_marshal_config.Dimg{}
	rubyDimg.DimgBase = c.DimgBase.ToRuby()

	if c.Shell != nil {
		rubyDimg.Shell = c.Shell.ToRuby()
	}

	if c.Docker != nil {
		rubyDimg.Docker = c.Docker.ToRuby()
	}
	rubyDimg.Docker.From = c.From

	return rubyDimg
}
