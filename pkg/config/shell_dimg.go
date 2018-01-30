package config

import (
	"github.com/flant/dapp/pkg/config/ruby_marshal_config"
)

type ShellDimg struct{ *ShellBase }

func (c *ShellDimg) Validate() error {
	return nil
}

func (c *ShellDimg) ToRuby() ruby_marshal_config.ShellDimg {
	shellDimg := ruby_marshal_config.ShellDimg{}
	shellDimg.BeforeInstall.Run = c.BeforeInstall
	shellDimg.Install.Run = c.Install
	shellDimg.BeforeSetup.Run = c.BeforeSetup
	shellDimg.Setup.Run = c.Setup
	return shellDimg
}
