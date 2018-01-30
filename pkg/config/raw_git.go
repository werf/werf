package config

type RawGit struct {
	RawGitExportBase     `yaml:",inline"`
	As                   string                `yaml:"as,omitempty"`
	Url                  string                `yaml:"url,omitempty"`
	Branch               string                `yaml:"branch,omitempty"`
	Commit               string                `yaml:"commit,omitempty"`
	RawStageDependencies *RawStageDependencies `yaml:"stageDependencies,omitempty"`

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *RawGit) Type() string {
	if c.Url != "" {
		return "remote"
	}
	return "local"
}

func (c *RawGit) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain RawGit
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	if err := CheckOverflow(c.UnsupportedAttributes, c); err != nil {
		return err
	}

	return nil
}

func (c *RawGit) ToGitLocalDirective() (gitLocal *GitLocal, err error) {
	gitLocal = &GitLocal{}

	if exportBase, err := c.RawGitExportBase.ToDirective(); err != nil {
		return nil, err
	} else {
		gitLocal.ExportBase = exportBase
	}

	if c.RawStageDependencies != nil {
		if stageDependencies, err := c.RawStageDependencies.ToDirective(); err != nil {
			return nil, err
		} else {
			gitLocal.StageDependencies = stageDependencies
		}
	}

	gitLocal.As = c.As

	return gitLocal, nil
}

func (c *RawGit) ToGitRemoteDirective() (gitRemote *GitRemote, err error) {
	gitRemote = &GitRemote{}

	if gitLocal, err := c.ToGitLocalDirective(); err != nil {
		return nil, err
	} else {
		gitRemote.GitLocal = gitLocal
	}

	gitRemote.Branch = c.Branch
	gitRemote.Commit = c.Commit
	gitRemote.Url = c.Url
	// TODO: gitRemote.Name = вычленить имя из c.Url

	if err := gitRemote.Validation(); err != nil {
		return nil, err
	}

	return gitRemote, nil
}
