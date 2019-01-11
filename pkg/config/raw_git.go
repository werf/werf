package config

import (
	"fmt"
	"regexp"
)

type rawGit struct {
	rawGitExport         `yaml:",inline"`
	As                   string                `yaml:"as,omitempty"`
	Url                  string                `yaml:"url,omitempty"`
	Branch               string                `yaml:"branch,omitempty"`
	Tag                  string                `yaml:"tag,omitempty"`
	Commit               string                `yaml:"commit,omitempty"`
	RawStageDependencies *rawStageDependencies `yaml:"stageDependencies,omitempty"`

	rawDimg *rawDimg `yaml:"-"` // parent

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *rawGit) configSection() interface{} {
	return c
}

func (c *rawGit) doc() *doc {
	return c.rawDimg.doc
}

func (c *rawGit) gitType() string {
	if c.Url != "" {
		return "remote"
	}
	return "local"
}

func (c *rawGit) UnmarshalYAML(unmarshal func(interface{}) error) error {
	c.rawGitExport = newRawGitExport()
	if parent, ok := parentStack.Peek().(*rawDimg); ok {
		c.rawDimg = parent
	}

	parentStack.Push(c)
	type plain rawGit
	err := unmarshal((*plain)(c))
	parentStack.Pop()
	if err != nil {
		return err
	}

	c.rawGitExport.inlinedIntoRaw(c)

	if err := checkOverflow(c.UnsupportedAttributes, c, c.rawDimg.doc); err != nil {
		return err
	}

	return nil
}

func (c *rawGit) toGitLocalDirective() (gitLocal *GitLocal, err error) {
	gitLocal = &GitLocal{}

	if gitLocalExport, err := c.toGitLocalExportDirective(); err != nil {
		return nil, err
	} else {
		gitLocal.GitLocalExport = gitLocalExport
	}

	gitLocal.As = c.As

	gitLocal.raw = c

	if err := c.validateGitLocalDirective(gitLocal); err != nil {
		return nil, err
	}

	return gitLocal, nil
}

func (c *rawGit) validateGitLocalDirective(gitLocal *GitLocal) (err error) {
	if c.Branch != "" || c.Commit != "" || c.Tag != "" {
		return newDetailedConfigError("specify `branch: BRANCH`, `tag: TAG` and `commit: COMMIT` only for remote git!", nil, c.rawDimg.doc)
	}

	if err := gitLocal.validate(); err != nil {
		return err
	}

	return nil
}

func (c *rawGit) toGitLocalExportDirective() (gitLocalExport *GitLocalExport, err error) {
	gitLocalExport = &GitLocalExport{}

	gitLocalExport.GitExportBase = &GitExportBase{}
	if gitExport, err := c.rawGitExport.toDirective(); err != nil {
		return nil, err
	} else {
		gitLocalExport.GitExportBase.GitExport = gitExport
	}

	if c.RawStageDependencies != nil {
		if stageDependencies, err := c.RawStageDependencies.toDirective(); err != nil {
			return nil, err
		} else {
			gitLocalExport.StageDependencies = stageDependencies
		}
	}

	gitLocalExport.raw = c

	if err := c.validateGitLocalExportDirective(gitLocalExport); err != nil {
		return nil, err
	}

	return gitLocalExport, nil
}

func (c *rawGit) validateGitLocalExportDirective(gitLocalExport *GitLocalExport) (err error) {
	if err := gitLocalExport.validate(); err != nil {
		return err
	}

	return nil
}

func (c *rawGit) toGitRemoteDirective() (gitRemote *GitRemote, err error) {
	gitRemote = &GitRemote{}

	if gitRemoteExport, err := c.toGitRemoteExportDirective(); err != nil {
		return nil, err
	} else {
		gitRemote.GitRemoteExport = gitRemoteExport
	}

	gitRemote.As = c.As
	gitRemote.Url = c.Url

	if url, err := c.getNameFromUrl(); err != nil {
		return nil, newDetailedConfigError(err.Error(), c, c.rawDimg.doc)
	} else {
		gitRemote.Name = url
	}

	gitRemote.raw = c

	if err := c.validateGitRemoteDirective(gitRemote); err != nil {
		return nil, err
	}

	return gitRemote, nil
}

func (c *rawGit) getNameFromUrl() (string, error) {
	return getGitName(c.Url)
}

func getGitName(remoteOriginUrl string) (string, error) {
	r := regexp.MustCompile(`.*?([^:/ ]+/[^/ ]+)\.git$`)
	match := r.FindStringSubmatch(remoteOriginUrl)
	if len(match) == 2 {
		return match[1], nil
	} else {
		return "", fmt.Errorf("cannot determine repo name from `url: %s`: url is not fit `.*?([^:/ ]+/[^/ ]+)\\.git$` regex", remoteOriginUrl)
	}
}

func (c *rawGit) validateGitRemoteDirective(gitRemote *GitRemote) (err error) {
	if err := gitRemote.validate(); err != nil {
		return err
	}

	return nil
}

func (c *rawGit) toGitRemoteExportDirective() (gitRemoteExport *GitRemoteExport, err error) {
	gitRemoteExport = &GitRemoteExport{}

	if gitLocalExport, err := c.toGitLocalExportDirective(); err != nil {
		return nil, err
	} else {
		gitRemoteExport.GitLocalExport = gitLocalExport
	}

	gitRemoteExport.Branch = c.Branch
	gitRemoteExport.Tag = c.Tag
	gitRemoteExport.Commit = c.Commit

	gitRemoteExport.raw = c

	if err := c.validateGitRemoteExportDirective(gitRemoteExport); err != nil {
		return nil, err
	}

	return gitRemoteExport, nil
}

func (c *rawGit) validateGitRemoteExportDirective(gitRemoteExport *GitRemoteExport) (err error) {
	if err := gitRemoteExport.validate(); err != nil {
		return err
	}

	return nil
}
