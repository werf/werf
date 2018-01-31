package config

import "github.com/flant/dapp/pkg/config/ruby_marshal_config"

type GitLocalExport struct {
	*GitExportBase

	Raw *RawGit
}

func (c *GitLocalExport) Validate() error {
	return nil
}

func (c *GitLocalExport) ToRuby() ruby_marshal_config.GitArtifactLocalExport {
	rubyGitArtifactLocalExport := ruby_marshal_config.GitArtifactLocalExport{}
	if c.ExportBase != nil {
		rubyGitArtifactLocalExport.ArtifactBaseExport = c.ExportBase.ToRuby()
	}
	if c.StageDependencies != nil {
		rubyGitArtifactLocalExport.StageDependencies = c.StageDependencies.ToRuby()
	}
	return rubyGitArtifactLocalExport
}
