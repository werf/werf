package config

import (
	"fmt"

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
	if !AllRelativePaths(c.Install) {
		return fmt.Errorf("`Install` should contain relative paths!") // FIXME
	} else if !AllRelativePaths(c.Setup) {
		return fmt.Errorf("`Setup` should contain relative paths!") // FIXME
	} else if !AllRelativePaths(c.BeforeSetup) {
		return fmt.Errorf("`BeforeSetup` should contain relative paths!") // FIXME
	} else if !AllRelativePaths(c.BuildArtifact) {
		return fmt.Errorf("`BuildArtifact` should contain relative paths!") // FIXME
	}
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
