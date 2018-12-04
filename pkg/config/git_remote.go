package config

import (
	"github.com/flant/dapp/pkg/config/ruby_marshal_config"
)

type GitRemote struct {
	*GitRemoteExport
	As   string
	Name string
	Url  string

	raw *rawGit
}

func (c *GitRemote) validate() error {
	return nil
}

func (c *GitRemote) toRuby() ruby_marshal_config.GitArtifactRemote {
	rubyGitArtifactRemote := ruby_marshal_config.GitArtifactRemote{}
	rubyGitArtifactRemote.Url = c.Url
	rubyGitArtifactRemote.Name = c.Name
	rubyGitArtifactRemote.As = c.As
	if c.GitRemoteExport != nil {
		rubyGitArtifactRemote.Export = append(rubyGitArtifactRemote.Export, c.GitRemoteExport.toRuby())
	}
	return rubyGitArtifactRemote
}
