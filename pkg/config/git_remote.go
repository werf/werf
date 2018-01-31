package config

import (
	"fmt"

	"github.com/flant/dapp/pkg/config/ruby_marshal_config"
)

type GitRemote struct {
	*GitLocal
	Name   string
	Branch string
	Commit string
	Url    string

	Raw *RawGit
}

func (c *GitRemote) Validate() error {
	if c.Branch != "" && c.Commit != "" {
		return fmt.Errorf("conflict between `Branch` && `Commit` directives") // FIXME
	}
	return nil
}

func (c *GitRemote) ToRuby() ruby_marshal_config.GitArtifactRemoteExport {
	rubyGitArtifactRemoteExport := ruby_marshal_config.GitArtifactRemoteExport{}
	rubyGitArtifactRemoteExport.GitArtifactLocalExport = c.GitLocal.ToRuby()
	rubyGitArtifactRemoteExport.Url = c.Url
	rubyGitArtifactRemoteExport.Branch = c.Branch
	rubyGitArtifactRemoteExport.Commit = c.Commit
	rubyGitArtifactRemoteExport.Name = c.Name
	return rubyGitArtifactRemoteExport
}
