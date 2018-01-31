package config

import "github.com/flant/dapp/pkg/config/ruby_marshal_config"

type GitLocal struct {
	*GitLocalExport
	As string

	Raw *RawGit
}

func (c *GitLocal) Validate() error {
	return nil
}

func (c *GitLocal) ToRuby() ruby_marshal_config.GitArtifactLocal {
	rubyGitArtifactLocal := ruby_marshal_config.GitArtifactLocal{}
	if c.GitLocalExport != nil {
		rubyGitArtifactLocal.Export = append(rubyGitArtifactLocal.Export, c.GitLocalExport.ToRuby())
	}
	rubyGitArtifactLocal.As = c.As
	return rubyGitArtifactLocal
}
