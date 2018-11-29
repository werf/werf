package config

import "github.com/flant/dapp/pkg/config/ruby_marshal_config"

type GitRemoteExport struct {
	*GitLocalExport
	Branch string
	Tag    string
	Commit string

	raw *rawGit
}

func (c *GitRemoteExport) validate() error {
	if !oneOrNone([]bool{c.Branch != "", c.Commit != "", c.Tag != ""}) {
		return newDetailedConfigError("Specify only `branch: BRANCH`, `tag: TAG` or `commit: COMMIT` for remote git!", c.raw, c.raw.rawDimg.doc)
	}
	return nil
}

func (c *GitRemoteExport) toRuby() ruby_marshal_config.GitArtifactRemoteExport {
	rubyGitArtifactRemoteExport := ruby_marshal_config.GitArtifactRemoteExport{}
	if c.GitLocalExport != nil {
		rubyGitArtifactRemoteExport.GitArtifactLocalExport = c.GitLocalExport.toRuby()
	}
	rubyGitArtifactRemoteExport.Branch = c.Branch
	rubyGitArtifactRemoteExport.Tag = c.Tag
	rubyGitArtifactRemoteExport.Commit = c.Commit
	return rubyGitArtifactRemoteExport
}
