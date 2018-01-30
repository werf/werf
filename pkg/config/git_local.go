package config

import "github.com/flant/dapp/pkg/config/ruby_marshal_config"

type GitLocal struct {
	*GitBase
	As                string
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
