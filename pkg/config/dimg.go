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
	if !OneOrNone([]bool{c.Shell != nil, c.Ansible != nil}) {
		return fmt.Errorf("Cannot use shell and ansible builders at the same time!\n\n%s", DumpConfigDoc(c.DimgBase.Raw.Doc))
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
