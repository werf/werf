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
		return fmt.Errorf("`install: [PATH, ...]|PATH` should be relative paths!\n\n%s\n%s", DumpConfigSection(c.Raw), DumpConfigDoc(c.Raw.RawGit.RawDimg.Doc))
	} else if !AllRelativePaths(c.Setup) {
		return fmt.Errorf("`setup: [PATH, ...]|PATH` should be relative paths!\n\n%s\n%s", DumpConfigSection(c.Raw), DumpConfigDoc(c.Raw.RawGit.RawDimg.Doc))
	} else if !AllRelativePaths(c.BeforeSetup) {
		return fmt.Errorf("`beforeSetup: [PATH, ...]|PATH` should be relative paths!\n\n%s\n%s", DumpConfigSection(c.Raw), DumpConfigDoc(c.Raw.RawGit.RawDimg.Doc))
	} else if !AllRelativePaths(c.BuildArtifact) {
		return fmt.Errorf("`buildArtifact: [PATH, ...]|PATH` should be relative paths!\n\n%s\n%s", DumpConfigSection(c.Raw), DumpConfigDoc(c.Raw.RawGit.RawDimg.Doc))
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
