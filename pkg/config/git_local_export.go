package config

import "github.com/flant/dapp/pkg/config/ruby_marshal_config"

type GitLocalExport struct {
	*GitExportBase

	raw *rawGit
}

func (c *GitLocalExport) validate() error {
	return nil
}

func (c *GitLocalExport) toRuby() ruby_marshal_config.GitArtifactLocalExport {
	rubyGitArtifactLocalExport := ruby_marshal_config.GitArtifactLocalExport{}
	if c.ExportBase != nil {
		rubyGitArtifactLocalExport.ArtifactBaseExport = c.ExportBase.toRuby()
	}
	if c.StageDependencies != nil {
		rubyGitArtifactLocalExport.StageDependencies = c.StageDependencies.toRuby()
	}
	return rubyGitArtifactLocalExport
}
