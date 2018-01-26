package config

import "github.com/flant/dapp/pkg/config/ruby_marshal_config"

type GitLocal struct {
	*ExportBase
	As                string
	StageDependencies *StageDependencies
}

type StageDependencies struct {
	Install       []string
	Setup         []string
	BeforeSetup   []string
	BuildArtifact []string
}

func (c *StageDependencies) Validation() error {
	// TODO: валидация относительных путей
	return nil
}

func (c *GitLocal) ToRuby() ruby_marshal_config.GitArtifactLocalExport {
	rubyGitArtifactLocalExport := ruby_marshal_config.GitArtifactLocalExport{}
	if c.ExportBase != nil {
		rubyGitArtifactLocalExport.ArtifactBaseExport = c.ExportBase.ToRuby()
	}
	if c.StageDependencies != nil {
		rubyGitArtifactLocalExport.StageDependencies = c.StageDependencies.ToRuby()
	}
	rubyGitArtifactLocalExport.As = c.As
	return rubyGitArtifactLocalExport
}

func (c *StageDependencies) ToRuby() ruby_marshal_config.StageDependencies {
	rubyStageDependencies := ruby_marshal_config.StageDependencies{}
	rubyStageDependencies.Install = c.Install
	rubyStageDependencies.BeforeSetup = c.BeforeSetup
	rubyStageDependencies.Setup = c.Setup
	rubyStageDependencies.BuildArtifact = c.BuildArtifact
	return rubyStageDependencies
}
