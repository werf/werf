package config

import "github.com/flant/dapp/pkg/config/ruby_marshal_config"

type GitRemoteExport struct {
	*GitLocalExport
	Branch string
	Tag    string
	Commit string

	Raw *RawGit
}

func (c *GitRemoteExport) Validate() error {
	if !OneOrNone([]bool{c.Branch != "", c.Commit != "", c.Tag != ""}) {
		return NewDetailedConfigError("Specify only `branch: BRANCH`, `tag: TAG` or `commit: COMMIT` for remote git!", c.Raw, c.Raw.RawDimg.Doc)
	}
	return nil
}

func (c *GitRemoteExport) ToRuby() ruby_marshal_config.GitArtifactRemoteExport {
	rubyGitArtifactRemoteExport := ruby_marshal_config.GitArtifactRemoteExport{}
	if c.GitLocalExport != nil {
		rubyGitArtifactRemoteExport.GitArtifactLocalExport = c.GitLocalExport.ToRuby()
	}
	rubyGitArtifactRemoteExport.Branch = c.Branch
	rubyGitArtifactRemoteExport.Tag = c.Tag
	rubyGitArtifactRemoteExport.Commit = c.Commit
	return rubyGitArtifactRemoteExport
}
