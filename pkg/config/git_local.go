package config

import "github.com/flant/dapp/pkg/config/ruby_marshal_config"

type GitLocal struct {
	*GitLocalExport
	As string

	raw *rawGit
}

func (c *GitLocal) validate() error {
	return nil
}

func (c *GitLocal) toRuby() ruby_marshal_config.GitArtifactLocal {
	rubyGitArtifactLocal := ruby_marshal_config.GitArtifactLocal{}
	if c.GitLocalExport != nil {
		rubyGitArtifactLocal.Export = append(rubyGitArtifactLocal.Export, c.GitLocalExport.toRuby())
	}
	rubyGitArtifactLocal.As = c.As
	return rubyGitArtifactLocal
}
