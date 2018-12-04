package config

import "github.com/flant/dapp/pkg/config/ruby_marshal_config"

type GitManager struct {
	Local  []*GitLocal
	Remote []*GitRemote
}

func (c *GitManager) toRuby() ruby_marshal_config.GitArtifact {
	gitArtifact := &ruby_marshal_config.GitArtifact{}

	for _, local := range c.Local {
		gitArtifact.Local = append(gitArtifact.Local, local.toRuby())
	}

	for _, remote := range c.Remote {
		gitArtifact.Remote = append(gitArtifact.Remote, remote.toRuby())
	}

	return *gitArtifact
}
