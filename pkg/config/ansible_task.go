package config

import (
	"github.com/flant/dapp/pkg/config/ruby_marshal_config"
)

type AnsibleTask struct {
	Config interface{}

	raw *rawAnsibleTask

	DumpConfigSection string // FIXME: reject in golang binary
}

func (c *AnsibleTask) validate() error {
	return nil
}

func (c *AnsibleTask) toRuby() ruby_marshal_config.AnsibleTask {
	rubyAnsibleTask := ruby_marshal_config.AnsibleTask{}
	rubyAnsibleTask.Config = c.Config
	rubyAnsibleTask.DumpConfigSection = dumpConfigSection(c.raw)
	return rubyAnsibleTask
}
