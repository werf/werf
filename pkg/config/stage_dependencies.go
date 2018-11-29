package config

import "github.com/flant/dapp/pkg/config/ruby_marshal_config"

type StageDependencies struct {
	Install       []string
	Setup         []string
	BeforeSetup   []string
	BuildArtifact []string

	raw *rawStageDependencies
}

func (c *StageDependencies) validate() error {
	if !allRelativePaths(c.Install) {
		return newDetailedConfigError("`install: [PATH, ...]|PATH` should be relative paths!", c.raw, c.raw.rawGit.rawDimg.doc)
	} else if !allRelativePaths(c.Setup) {
		return newDetailedConfigError("`setup: [PATH, ...]|PATH` should be relative paths!", c.raw, c.raw.rawGit.rawDimg.doc)
	} else if !allRelativePaths(c.BeforeSetup) {
		return newDetailedConfigError("`beforeSetup: [PATH, ...]|PATH` should be relative paths!", c.raw, c.raw.rawGit.rawDimg.doc)
	} else if !allRelativePaths(c.BuildArtifact) {
		return newDetailedConfigError("`buildArtifact: [PATH, ...]|PATH` should be relative paths!", c.raw, c.raw.rawGit.rawDimg.doc)
	}
	return nil
}

func (c *StageDependencies) toRuby() ruby_marshal_config.StageDependencies {
	rubyStageDependencies := ruby_marshal_config.StageDependencies{}
	rubyStageDependencies.Install = c.Install
	rubyStageDependencies.BeforeSetup = c.BeforeSetup
	rubyStageDependencies.Setup = c.Setup
	rubyStageDependencies.BuildArtifact = c.BuildArtifact
	return rubyStageDependencies
}
