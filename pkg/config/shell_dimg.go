package config

import (
	"github.com/flant/dapp/pkg/config/ruby_marshal_config"
)

type ShellDimg struct{ *ShellBase }

func (c *ShellDimg) validate() error {
	return nil
}

func (c *ShellDimg) toRuby() ruby_marshal_config.ShellDimg {
	shellDimg := ruby_marshal_config.ShellDimg{}

	shellDimg.Version = c.CacheVersion
	shellDimg.BeforeInstall.Version = c.BeforeInstallCacheVersion
	shellDimg.BeforeInstall.Run = c.BeforeInstall
	shellDimg.Install.Version = c.InstallCacheVersion
	shellDimg.Install.Run = c.Install
	shellDimg.BeforeSetup.Version = c.BeforeSetupCacheVersion
	shellDimg.BeforeSetup.Run = c.BeforeSetup
	shellDimg.Setup.Version = c.SetupCacheVersion
	shellDimg.Setup.Run = c.Setup

	return shellDimg
}
