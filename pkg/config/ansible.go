package config

import (
	"github.com/flant/dapp/pkg/config/ruby_marshal_config"
)

type Ansible struct {
	BeforeInstall []*AnsibleTask
	Install       []*AnsibleTask
	BeforeSetup   []*AnsibleTask
	Setup         []*AnsibleTask

	Raw *RawAnsible
}

func (c *Ansible) Validate() error {
	return nil
}

func (c *Ansible) ToRuby() ruby_marshal_config.Ansible {
	rubyAnsible := ruby_marshal_config.Ansible{}

	for _, ansibleTask := range c.BeforeInstall {
		rubyAnsible.BeforeInstall = append(rubyAnsible.BeforeInstall, ansibleTask.ToRuby())
	}

	for _, ansibleTask := range c.Install {
		rubyAnsible.Install = append(rubyAnsible.Install, ansibleTask.ToRuby())
	}

	for _, ansibleTask := range c.BeforeSetup {
		rubyAnsible.BeforeSetup = append(rubyAnsible.BeforeSetup, ansibleTask.ToRuby())
	}

	for _, ansibleTask := range c.Setup {
		rubyAnsible.Setup = append(rubyAnsible.Setup, ansibleTask.ToRuby())
	}

	rubyAnsible.DumpConfigDoc = DumpConfigDoc(c.Raw.RawDimg.Doc)

	return rubyAnsible
}
