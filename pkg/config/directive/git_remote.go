package config

import "github.com/flant/dapp/pkg/config/ruby_marshal_config"

type GitRemote struct {
	*GitLocal
	// TODO: Name string
	Branch string
	Commit string
	Url    string
}

func (c *GitLocal) Validation() error {
	// TODO: валидация одновременного использования `Branch` и `Commit`
	return nil
}

func (c *GitRemote) ToRuby() ruby_marshal_config.GitArtifactRemoteExport {
	rubyGitArtifactRemoteExport := ruby_marshal_config.GitArtifactRemoteExport{}
	rubyGitArtifactRemoteExport.GitArtifactLocalExport = c.GitLocal.ToRuby()
	rubyGitArtifactRemoteExport.Url = c.Url
	rubyGitArtifactRemoteExport.Branch = c.Branch
	rubyGitArtifactRemoteExport.Commit = c.Commit
	// TODO: rubyGitArtifactRemoteExport.Name = c.Name
	return rubyGitArtifactRemoteExport
}
