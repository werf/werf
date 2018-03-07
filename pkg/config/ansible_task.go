package config

import (
	"github.com/flant/dapp/pkg/config/ruby_marshal_config"
)

type AnsibleTask struct {
	Config interface{}

	Raw *RawAnsibleTask
}

func (c *AnsibleTask) Validate() error {
	return nil
}

func (c *AnsibleTask) ToRuby() ruby_marshal_config.AnsibleTask {
	rubyAnsibleTask := ruby_marshal_config.AnsibleTask{}
	rubyAnsibleTask.Config = c.Config
	rubyAnsibleTask.DumpConfigSection = DumpConfigSection(c.Raw)
	return rubyAnsibleTask
}
