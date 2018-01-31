package config

import (
	"fmt"
	"regexp"
)

type RawGit struct {
	RawGitExport         `yaml:",inline"`
	As                   string                `yaml:"as,omitempty"`
	Url                  string                `yaml:"url,omitempty"`
	Branch               string                `yaml:"branch,omitempty"`
	Commit               string                `yaml:"commit,omitempty"`
	RawStageDependencies *RawStageDependencies `yaml:"stageDependencies,omitempty"`

	RawDimg *RawDimg `yaml:"-"` // parent

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *RawGit) ConfigSection() interface{} {
	return c
}

func (c *RawGit) Doc() *Doc {
	return c.RawDimg.Doc
}

func (c *RawGit) Type() string {
	if c.Url != "" {
		return "remote"
	}
	return "local"
}

func (c *RawGit) UnmarshalYAML(unmarshal func(interface{}) error) error {
	c.RawGitExport.RawExportBase = NewRawExportBase()
	if parent, ok := ParentStack.Peek().(*RawDimg); ok {
		c.RawDimg = parent
	}

	ParentStack.Push(c)
	type plain RawGit
	err := unmarshal((*plain)(c))
	ParentStack.Pop()
	if err != nil {
		return err
	}

	c.RawGitExport.InlinedIntoRaw(c)

	if err := CheckOverflow(c.UnsupportedAttributes, c); err != nil {
		return err
	}

	return nil
}

func (c *RawGit) ToGitLocalDirective() (gitLocal *GitLocal, err error) {
	gitLocal = &GitLocal{}
	gitLocal.GitBase = &GitBase{}

	if gitExport, err := c.RawGitExport.ToDirective(); err != nil {
		return nil, err
	} else {
		gitLocal.GitBase.GitExport = gitExport
	}

	if c.RawStageDependencies != nil {
		if stageDependencies, err := c.RawStageDependencies.ToDirective(); err != nil {
			return nil, err
		} else {
			gitLocal.StageDependencies = stageDependencies
		}
	}

	gitLocal.As = c.As

	gitLocal.Raw = c

	if err := c.ValidateGitLocalDirective(gitLocal); err != nil {
		return nil, err
	}

	return gitLocal, nil
}

func (c *RawGit) ValidateGitLocalDirective(gitLocal *GitLocal) (err error) {
	if err := gitLocal.Validate(); err != nil {
		return err
	}

	return nil
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

	r := regexp.MustCompile(`.*?([^/ ]+/[^/ ]+)(\.git)?`)
	match := r.FindStringSubmatch(c.Url)
	if len(match) == 3 {
		gitRemote.Name = match[1]
	} else {
		return nil, fmt.Errorf("не удалось вычленить имя из `%s`", c.Url) // FIXME
	}

	gitRemote.Raw = c

	if err := c.ValidateGitRemoteDirective(gitRemote); err != nil {
		return nil, err
	}

	return gitRemote, nil
}

func (c *RawGit) ValidateGitRemoteDirective(gitRemote *GitRemote) (err error) {
	if err := gitRemote.Validate(); err != nil {
		return err
	}

	return nil
}
