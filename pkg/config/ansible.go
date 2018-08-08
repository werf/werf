package config

import (
	"github.com/flant/dapp/pkg/config/ruby_marshal_config"
)

type Ansible struct {
	BeforeInstall             []*AnsibleTask
	Install                   []*AnsibleTask
	BeforeSetup               []*AnsibleTask
	Setup                     []*AnsibleTask
	BuildArtifact             []*AnsibleTask
	CacheVersion              string
	BeforeInstallCacheVersion string
	InstallCacheVersion       string
	BeforeSetupCacheVersion   string
	SetupCacheVersion         string
	BuildArtifactCacheVersion string

	Raw *RawAnsible

	DumpConfigSection string // FIXME: reject after a complete transition from ruby to golang
}

func (c *Ansible) Validate() error {
	return nil
}

func (c *Ansible) ToRuby() ruby_marshal_config.Ansible {
	rubyAnsible := ruby_marshal_config.Ansible{}

	rubyAnsible.Version = c.CacheVersion
	rubyAnsible.BeforeInstallVersion = c.BeforeInstallCacheVersion
	rubyAnsible.InstallVersion = c.InstallCacheVersion
	rubyAnsible.BeforeSetupVersion = c.BeforeSetupCacheVersion
	rubyAnsible.SetupVersion = c.SetupCacheVersion
	rubyAnsible.BuildArtifactVersion = c.BuildArtifactCacheVersion

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

	for _, ansibleTask := range c.BuildArtifact {
		rubyAnsible.BuildArtifact = append(rubyAnsible.BuildArtifact, ansibleTask.ToRuby())
	}

	rubyAnsible.DumpConfigDoc = DumpConfigDoc(c.Raw.RawDimg.Doc)

	return rubyAnsible
}
