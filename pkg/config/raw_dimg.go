package config

import (
	"fmt"
)

type RawDimg struct {
	Dimg      interface{}          `yaml:"-"`
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

func (c *RawDimg) SetAndValidateDimg() error {
	value, ok := c.UnsupportedAttributes["dimg"]
	if ok {
		delete(c.UnsupportedAttributes, "dimg")

		switch t := value.(type) {
		case string: // TODO: поддержка нескольких имён []string
			c.Dimg = value
		case nil:
			c.Dimg = ""
		default:
			return fmt.Errorf("Invalid dimg name `%v`!\n\n%s", t, DumpConfigDoc(c.Doc))
		}
	}

	return nil
}

func (c *RawDimg) UnmarshalYAML(unmarshal func(interface{}) error) error {
	ParentStack.Push(c)
	type plain RawDimg
	err := unmarshal((*plain)(c))
	ParentStack.Pop()
	if err != nil {
		return err
	}

	if err := c.SetAndValidateDimg(); err != nil {
		return err
	}

	if err := CheckOverflow(c.UnsupportedAttributes, nil, c.Doc); err != nil {
		return err
	}

	if err := c.ValidateType(); err != nil {
		return err
	}

	return nil
}

func (c *RawDimg) ValidateType() error {
	isDimg := c.Dimg != nil
	isArtifact := c.Artifact != ""

	if isDimg && isArtifact {
		return fmt.Errorf("Unknown doc type: one and only one of `dimg: NAME` or `artifact: NAME` non-empty name required!\n\n%s", DumpConfigDoc(c.Doc))
	} else if !(isDimg || isArtifact) {
		return fmt.Errorf("Unknown doc type: one of `dimg: NAME` or `artifact: NAME` non-empty name required!\n\n%s", DumpConfigDoc(c.Doc))
	}

	return nil
}

func (c *RawDimg) Type() string {
	if c.Dimg != nil {
		return "dimg"
	} else if c.Artifact != "" {
		return "artifact"
	}

	return ""
}

func (c *RawDimg) ToDirective() (dimg *Dimg, err error) {
	dimg = &Dimg{}

	if dimgBase, err := c.ToBaseDirective(c.Dimg.(string)); err != nil {
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
		return fmt.Errorf("`docker` section is not supported for artifact!\n\n%s", DumpConfigDoc(c.Doc))
	}

	if err := dimgArtifact.Validate(); err != nil {
		return err
	}

	return nil
}

func (c *RawDimg) ToBaseDirective(name string) (dimgBase *DimgBase, err error) {
	dimgBase = &DimgBase{}
	dimgBase.Name = name
	dimgBase.Bulder = "none"
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
