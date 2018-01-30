package config

import (
	"fmt"
)

type RawDimg struct {
	Dimg      string               `yaml:"dimg,omitempty"` // TODO: поддержка нескольких имён
	Artifact  string               `yaml:"artifact,omitempty"`
	From      string               `yaml:"from,omitempty"`
	RawGit    []*RawGit            `yaml:"git,omitempty"`
	RawShell  *RawShell            `yaml:"shell,omitempty"`
	RawChef   *RawChef             `yaml:"chef,omitempty"`
	RawMount  []*RawMount          `yaml:"mount,omitempty"`
	RawDocker *RawDocker           `yaml:"docker,omitempty"`
	RawImport []*RawArtifactImport `yaml:"import,omitempty"`

	Doc *Doc `yaml:"-"` // parent

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *RawDimg) UnmarshalYAML(unmarshal func(interface{}) error) error {
	ParentStack.Push(c)
	type plain RawDimg
	err := unmarshal((*plain)(c))
	ParentStack.Pop()
	if err != nil {
		return err
	}
	if err := CheckOverflow(c.UnsupportedAttributes, c); err != nil {
		return err
	}

	if err := c.ValidateType(); err != nil {
		return err
	}

	return nil
}

func (c *RawDimg) ValidateType() error {
	isDimg := c.Dimg != ""
	isArtifact := c.Artifact != ""

	if isDimg && isArtifact {
		return fmt.Errorf("объект не может быть и артефактом и dimg-ом одновременно!") // FIXME
	} else if !(isDimg || isArtifact) {
		return fmt.Errorf("объект не связан ни с dimg-ом ни с артефактом") // FIXME
	}

	return nil
}

func (c *RawDimg) Type() string {
	if c.Dimg != "" {
		return "dimg"
	} else if c.Artifact != "" {
		return "artifact"
	}

	return ""
}

func (c *RawDimg) ToDirective() (dimg *Dimg, err error) {
	dimg = &Dimg{}

	if dimgBase, err := c.ToBaseDirective(c.Dimg); err != nil {
		return nil, err
	} else {
		dimg.DimgBase = dimgBase
	}

	if c.RawShell != nil {
		dimg.Bulder = "shell"
		if shell, err := c.RawShell.ToDirective(); err != nil {
			return nil, err
		} else {
			dimg.Shell = shell
		}
	}

	if c.RawDocker != nil {
		if docker, err := c.RawDocker.ToDirective(); err != nil {
			return nil, err
		} else {
			dimg.Docker = docker
		}
	}

	if err := c.ValidateDirective(dimg); err != nil {
		return nil, err
	}

	return dimg, nil
}

func (c *RawDimg) ValidateDirective(dimg *Dimg) (err error) {
	if err := dimg.Validate(); err != nil {
		return err
	}

	return nil
}

func (c *RawDimg) ToArtifactDirective() (dimgArtifact *DimgArtifact, err error) {
	dimgArtifact = &DimgArtifact{}

	if dimgArtifact.DimgBase, err = c.ToBaseDirective(c.Artifact); err != nil {
		return nil, err
	}

	if c.RawShell != nil {
		dimgArtifact.Bulder = "shell"
		if dimgArtifact.Shell, err = c.RawShell.ToArtifactDirective(); err != nil {
			return nil, err
		}
	}

	if err := c.ValidateArtifactDirective(dimgArtifact); err != nil {
		return nil, err
	}

	return dimgArtifact, nil
}

func (c *RawDimg) ValidateArtifactDirective(dimgArtifact *DimgArtifact) (err error) {
	if c.RawDocker != nil {
		return fmt.Errorf("docker не поддерживается для артефакта!") // FIXME
	}

	if err := dimgArtifact.Validate(); err != nil {
		return err
	}

	return nil
}

func (c *RawDimg) ToBaseDirective(name string) (dimgBase *DimgBase, err error) {
	dimgBase = &DimgBase{}
	dimgBase.Name = name
	dimgBase.From = c.From

	dimgBase.Git = &GitManager{}
	for _, git := range c.RawGit {
		if git.Type() == "local" {
			if gitLocal, err := git.ToGitLocalDirective(); err != nil {
				return nil, err
			} else {
				dimgBase.Git.Local = append(dimgBase.Git.Local, gitLocal)
			}
		} else {
			if gitRemote, err := git.ToGitRemoteDirective(); err != nil {
				return nil, err
			} else {
				dimgBase.Git.Remote = append(dimgBase.Git.Remote, gitRemote)
			}
		}
	}

	if c.RawChef != nil {
		dimgBase.Bulder = "chef"
		if dimgBase.Chef, err = c.RawChef.ToDirective(); err != nil {
			return nil, err
		}
	}

	for _, mount := range c.RawMount {
		if dimgMount, err := mount.ToDirective(); err != nil {
			return nil, err
		} else {
			dimgBase.Mount = append(dimgBase.Mount, dimgMount)
		}
	}

	for _, importArtifact := range c.RawImport {
		if importArtifactDirective, err := importArtifact.ToDirective(); err != nil {
			return nil, err
		} else {
			dimgBase.Import = append(dimgBase.Import, importArtifactDirective)
		}
	}

	dimgBase.Raw = c

	if err := c.ValidateBaseDirective(dimgBase); err != nil {
		return nil, err
	}

	return dimgBase, nil
}

func (c *RawDimg) ValidateBaseDirective(dimgBase *DimgBase) (err error) {
	if err := dimgBase.Validate(); err != nil {
		return err
	}

	return nil
}
