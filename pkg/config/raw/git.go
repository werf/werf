package config

import "github.com/flant/dapp/pkg/config/directive"

type GitBase struct {
	dimg *Dimg

	ExportBase        `yaml:",inline"`
	As                string             `yaml:"as,omitempty"`
	Url               string             `yaml:"url,omitempty"`
	Branch            string             `yaml:"branch,omitempty"`
	Commit            string             `yaml:"commit,omitempty"`
	StageDependencies *StageDependencies `yaml:"stageDependencies,omitempty"`

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *GitBase) Type() string {
	if c.Url != "" {
		return "remote"
	}
	return "local"
}

func (c *GitBase) UnmarshalYAML(unmarshal func(interface{}) error) error {
	c.dimg = YamlParseContext[-1].(*Dimg)

	YamlParseContext = append(YamlParseContext, c)

	type plain GitBase
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	if err := CheckOverflow(c.UnsupportedAttributes, c); err != nil {
		return err
	}

	YamlParseContext = YamlParseContext[:len(YamlParseContext)-1]

	return nil
}

type StageDependencies struct {
	Install       interface{} `yaml:"install,omitempty"`
	Setup         interface{} `yaml:"setup,omitempty"`
	BeforeSetup   interface{} `yaml:"beforeSetup,omitempty"`
	BuildArtifact interface{} `yaml:"buildArtifact,omitempty"`

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *StageDependencies) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain StageDependencies
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	if err := CheckOverflow(c.UnsupportedAttributes, c); err != nil {
		return err
	}

	return nil
}

func (c *StageDependencies) ToDirective() (stageDependencies *config.StageDependencies, err error) {
	stageDependencies = &config.StageDependencies{}

	if install, err := InterfaceToStringArray(c.Install); err != nil {
		return nil, err
	} else {
		stageDependencies.Install = install
	}

	if beforeSetup, err := InterfaceToStringArray(c.BeforeSetup); err != nil {
		return nil, err
	} else {
		stageDependencies.BeforeSetup = beforeSetup
	}

	if setup, err := InterfaceToStringArray(c.Setup); err != nil {
		return nil, err
	} else {
		stageDependencies.Setup = setup
	}

	if buildArtifact, err := InterfaceToStringArray(c.BuildArtifact); err != nil {
		return nil, err
	} else {
		stageDependencies.BuildArtifact = buildArtifact
	}

	if err := stageDependencies.Validation(); err != nil {
		return nil, err
	}

	return stageDependencies, nil
}

func (c *GitBase) ToGitLocalDirective() (gitLocal *config.GitLocal, err error) {
	gitLocal = &config.GitLocal{}

	if exportBase, err := c.ExportBase.ToDirective(); err != nil {
		return nil, err
	} else {
		gitLocal.ExportBase = exportBase
	}

	if c.StageDependencies != nil {
		if stageDependencies, err := c.StageDependencies.ToDirective(); err != nil {
			return nil, err
		} else {
			gitLocal.StageDependencies = stageDependencies
		}
	}

	gitLocal.As = c.As

	return gitLocal, nil
}

func (c *GitBase) ToGitRemoteDirective() (gitRemote *config.GitRemote, err error) {
	gitRemote = &config.GitRemote{}

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
