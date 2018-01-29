package config

import "github.com/flant/dapp/pkg/config/ruby_marshal_config"

type GitManager struct {
	Local  []*GitLocal
	Remote []*GitRemote
}

func (c *GitManager) ToRuby() ruby_marshal_config.GitArtifact {
	rubyGitArtifactLocal := ruby_marshal_config.GitArtifactLocal{}
	for _, local := range c.Local {
		rubyGitArtifactLocal.Export = append(rubyGitArtifactLocal.Export, local.ToRuby())
	}

	rubyGitArtifactRemote := ruby_marshal_config.GitArtifactRemote{}
	for _, remote := range c.Remote {
		rubyGitArtifactRemote.Export = append(rubyGitArtifactRemote.Export, remote.ToRuby())
	}

	return ruby_marshal_config.GitArtifact{
		Local:  []ruby_marshal_config.GitArtifactLocal{rubyGitArtifactLocal},
		Remote: []ruby_marshal_config.GitArtifactRemote{rubyGitArtifactRemote},
	}
}
