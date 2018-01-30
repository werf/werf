package config

import (
	"github.com/flant/dapp/pkg/config/ruby_marshal_config"
)

type StageDependencies struct {
	Install       []string
	Setup         []string
	BeforeSetup   []string
	BuildArtifact []string

	Raw *RawStageDependencies
}

func (c *StageDependencies) Validate() error {
	// TODO: валидация относительных путей
	return nil
}

func (c *StageDependencies) ToRuby() ruby_marshal_config.StageDependencies {
	rubyStageDependencies := ruby_marshal_config.StageDependencies{}
	rubyStageDependencies.Install = c.Install
	rubyStageDependencies.BeforeSetup = c.BeforeSetup
	rubyStageDependencies.Setup = c.Setup
	rubyStageDependencies.BuildArtifact = c.BuildArtifact
	return rubyStageDependencies
}
