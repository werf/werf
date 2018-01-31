package config

import "github.com/flant/dapp/pkg/config/ruby_marshal_config"

type GitManager struct {
	Local  []*GitLocal
	Remote []*GitRemote
}

func (c *GitManager) ToRuby() ruby_marshal_config.GitArtifact {
	gitArtifact := &ruby_marshal_config.GitArtifact{}

	for _, local := range c.Local {
		gitArtifact.Local = append(gitArtifact.Local, local.ToRuby())
	}

	for _, remote := range c.Remote {
		gitArtifact.Remote = append(gitArtifact.Remote, remote.ToRuby())
	}

	return *gitArtifact
}
