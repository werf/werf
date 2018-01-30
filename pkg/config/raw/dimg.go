package config

import (
	"fmt"
	"github.com/flant/dapp/pkg/config/directive"
)

type Dimg struct {
	Dimg     string            `yaml:"dimg,omitempty"` // TODO: поддержка нескольких имён
	Artifact string            `yaml:"artifact,omitempty"`
	From     string            `yaml:"from,omitempty"`
	Git      []*GitBase        `yaml:"git,omitempty"`
	Shell    *Shell            `yaml:"shell,omitempty"`
	Chef     *Chef             `yaml:"chef,omitempty"`
	Mount    []*Mount          `yaml:"mount,omitempty"`
	Docker   *Docker           `yaml:"docker,omitempty"`
	Import   []*ArtifactImport `yaml:"import,omitempty"`

	Doc      *Doc

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *Dimg) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Dimg
	if err := unmarshal((*plain)(c)); err != nil {
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

func (c *Dimg) ValidateType() error {
	isDimg := c.Dimg != ""
	isArtifact := c.Artifact != ""

	if isDimg && isArtifact {
		return fmt.Errorf("объект не может быть и артефактом и dimg-ом одновременно!") // FIXME
	} else if !(isDimg || isArtifact) {
		return fmt.Errorf("объект не связан ни с dimg-ом ни с артефактом") // FIXME
	}

	return nil
}

func (c *Dimg) Type() string {
	if c.Dimg != "" {
		return "dimg"
	} else if c.Artifact != "" {
		return "artifact"
	}

	return ""
}

func (c *Dimg) ToDimg() (dimg *config.Dimg, err error) {
	dimg = &config.Dimg{}

	if dimgBase, err := c.ToDimgBase(c.Dimg); err != nil {
		return nil, err
	} else {
		dimg.DimgBase = dimgBase
	}

	if c.Shell != nil {
		dimg.Bulder = "shell"
		if shell, err := c.Shell.ToDirective(); err != nil {
			return nil, err
		} else {
			dimg.Shell = shell
		}
	}

	if c.Docker != nil {
		if docker, err := c.Docker.ToDirective(); err != nil {
			return nil, err
		} else {
			dimg.Docker = docker
		}
	}

	if err := dimg.Validate(); err != nil {
		return nil, err
	}

	return dimg, nil
}

func (c *Dimg) ToDimgArtifact() (dimgArtifact *config.DimgArtifact, err error) {
	dimgArtifact = &config.DimgArtifact{}

	if dimgArtifact.DimgBase, err = c.ToDimgBase(c.Artifact); err != nil {
		return nil, err
	}

	if c.Shell != nil {
		dimgArtifact.Bulder = "shell"
		if dimgArtifact.Shell, err = c.Shell.ToArtifact(); err != nil {
			return nil, err
		}
	}

	if c.Docker != nil {
		return nil, fmt.Errorf("docker не поддерживается для артефакта!") // FIXME
	}

	if err := dimgArtifact.Validate(); err != nil {
		return nil, err
	}

	return dimgArtifact, nil
}

func (c *Dimg) ToDimgBase(name string) (dimgBase *config.DimgBase, err error) {
	dimgBase = &config.DimgBase{}
	dimgBase.Name = name
	dimgBase.From = c.From

	dimgBase.Git = &config.GitManager{}
	for _, git := range c.Git {
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

	if c.Chef != nil {
		dimgBase.Bulder = "chef"
		if dimgBase.Chef, err = c.Chef.ToDirective(); err != nil {
			return nil, err
		}
	}

	for _, mount := range c.Mount {
		if dimgMount, err := mount.ToDirective(); err != nil {
			return nil, err
		} else {
			dimgBase.Mount = append(dimgBase.Mount, dimgMount)
		}
	}

	for _, importArtifact := range c.Import {
		if importArtifactDirective, err := importArtifact.ToDirective(); err != nil {
			return nil, err
		} else {
			dimgBase.Import = append(dimgBase.Import, importArtifactDirective)
		}
	}

	return dimgBase, nil
}
