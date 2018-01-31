package config

import (
	"fmt"

	"github.com/flant/dapp/pkg/config/ruby_marshal_config"
)

type GitRemoteExport struct {
	*GitLocalExport
	Branch string
	Commit string

	Raw *RawGit
}

func (c *GitRemoteExport) Validate() error {
	if c.Branch != "" && c.Commit != "" {
		return fmt.Errorf("Specify only `branch: BRANCH` or `commit: COMMIT` for git!\n\n%s\n%s", DumpConfigSection(c.Raw), DumpConfigDoc(c.Raw.RawDimg.Doc))
	}
	return nil
}

func (c *GitRemoteExport) ToRuby() ruby_marshal_config.GitArtifactRemoteExport {
	rubyGitArtifactRemoteExport := ruby_marshal_config.GitArtifactRemoteExport{}
	if c.GitLocalExport != nil {
		rubyGitArtifactRemoteExport.GitArtifactLocalExport = c.GitLocalExport.ToRuby()
	}
	rubyGitArtifactRemoteExport.Branch = c.Branch
	rubyGitArtifactRemoteExport.Commit = c.Commit
	return rubyGitArtifactRemoteExport
}
