package config

import (
	"github.com/flant/dapp/pkg/config/ruby_marshal_config"
)

type Ansible struct {
	BeforeInstall []interface{}
	Install       []interface{}
	BeforeSetup   []interface{}
	Setup         []interface{}

	Raw *RawAnsible
}

func (c *Ansible) Validate() error {
	return nil
}

func (c *Ansible) ToRuby() ruby_marshal_config.Ansible {
	rubyAnsible := ruby_marshal_config.Ansible{}
	rubyAnsible.BeforeInstall = c.BeforeInstall
	rubyAnsible.Install = c.Install
	rubyAnsible.BeforeSetup = c.BeforeSetup
	rubyAnsible.Setup = c.Setup
	return rubyAnsible
}
