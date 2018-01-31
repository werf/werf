package config

import (
	"fmt"

	"github.com/flant/dapp/pkg/config/ruby_marshal_config"
)

type GitRemote struct {
	*GitRemoteExport
	As   string
	Name string
	Url  string

	Raw *RawGit
}

func (c *GitRemote) Validate() error {
	if c.Branch != "" && c.Commit != "" {
		return fmt.Errorf("`branch` and `commit` fields cannot be used at the same time!\n\n%s\n%s", DumpConfigSection(c.Raw), DumpConfigDoc(c.Raw.RawDimg.Doc))
	}
	return nil
}

func (c *GitRemote) ToRuby() ruby_marshal_config.GitArtifactRemote {
	rubyGitArtifactRemote := ruby_marshal_config.GitArtifactRemote{}
	rubyGitArtifactRemote.Url = c.Url
	rubyGitArtifactRemote.Name = c.Name
	rubyGitArtifactRemote.As = c.As
	if c.GitRemoteExport != nil {
		rubyGitArtifactRemote.Export = append(rubyGitArtifactRemote.Export, c.GitRemoteExport.ToRuby())
	}
	return rubyGitArtifactRemote
}
