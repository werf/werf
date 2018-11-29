package config

import (
	"fmt"
	"strings"
)

type rawDimg struct {
	Dimgs            []string             `yaml:"-"`
	Artifact         string               `yaml:"artifact,omitempty"`
	From             string               `yaml:"from,omitempty"`
	FromCacheVersion string               `yaml:"fromCacheVersion,omitempty"`
	FromDimg         string               `yaml:"fromDimg,omitempty"`
	FromDimgArtifact string               `yaml:"fromDimgArtifact,omitempty"`
	RawGit           []*rawGit            `yaml:"git,omitempty"`
	RawShell         *rawShell            `yaml:"shell,omitempty"`
	RawAnsible       *rawAnsible          `yaml:"ansible,omitempty"`
	RawMount         []*rawMount          `yaml:"mount,omitempty"`
	RawDocker        *rawDocker           `yaml:"docker,omitempty"`
	RawImport        []*rawArtifactImport `yaml:"import,omitempty"`
	AsLayers         bool                 `yaml:"asLayers,omitempty"`

	doc *doc `yaml:"-"` // parent

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *rawDimg) setAndvalidateDimg() error {
	value, ok := c.UnsupportedAttributes["dimg"]
	if ok {
		delete(c.UnsupportedAttributes, "dimg")

		switch t := value.(type) {
		case []interface{}:
			if dimgs, err := InterfaceToStringArray(value, nil, c.doc); err != nil {
				return err
			} else {
				c.Dimgs = dimgs
			}
		case string:
			c.Dimgs = []string{value.(string)}
		case nil:
			c.Dimgs = []string{""}
		default:
			return newDetailedConfigError(fmt.Sprintf("Invalid dimg name `%v`!", t), nil, c.doc)
		}
	}

	return nil
}

func (c *rawDimg) UnmarshalYAML(unmarshal func(interface{}) error) error {
	parentStack.Push(c)
	type plain rawDimg
	err := unmarshal((*plain)(c))
	parentStack.Pop()
	if err != nil {
		return err
	}

	if err := c.setAndvalidateDimg(); err != nil {
		return err
	}

	if err := checkOverflow(c.UnsupportedAttributes, nil, c.doc); err != nil {
		return err
	}

	if err := c.validateDimgType(); err != nil {
		return err
	}

	return nil
}

func (c *rawDimg) validateDimgType() error {
	isDimg := len(c.Dimgs) != 0
	isArtifact := c.Artifact != ""

	if isDimg && isArtifact {
		return newDetailedConfigError("Unknown doc type: one and only one of `dimg: NAME` or `artifact: NAME` non-empty name required!", nil, c.doc)
	} else if !(isDimg || isArtifact) {
		return newDetailedConfigError("Unknown doc type: one of `dimg: NAME` or `artifact: NAME` non-empty name required!", nil, c.doc)
	}

	return nil
}

func (c *rawDimg) dimgType() string {
	if len(c.Dimgs) != 0 {
		return "dimgs"
	} else if c.Artifact != "" {
		return "artifact"
	}

	return ""
}

func (c *rawDimg) toDimgDirectives() (dimgs []*Dimg, err error) {
	for _, dimgName := range c.Dimgs {
		if dimg, err := c.toDimgDirective(dimgName); err != nil {
			return nil, err
		} else {
			dimgs = append(dimgs, dimg)
		}
	}

	return dimgs, nil
}

func (c *rawDimg) toDimgArtifactDirective() (dimgArtifact *DimgArtifact, err error) {
	dimgArtifact = &DimgArtifact{}
	if c.AsLayers {
		if dimgArtifact, err = c.toDimgArtifactAsLayersDirective(); err != nil {
			return nil, err
		}
	} else {
		if dimgArtifact.DimgBase, err = c.toDimgBaseDirective(c.Artifact); err != nil {
			return nil, err
		}

		if c.RawShell != nil {
			dimgArtifact.builder = "shell"
			if dimgArtifact.Shell, err = c.RawShell.toArtifactDirective(); err != nil {
				return nil, err
			}
		}
	}

	if err := c.validateArtifactDimgDirective(dimgArtifact); err != nil {
		return nil, err
	}

	return dimgArtifact, nil
}

func (c *rawDimg) toDimgDirective(name string) (dimg *Dimg, err error) {
	dimg = &Dimg{}

	if c.AsLayers {
		if dimg, err = c.toDimgAsLayersDirective(name); err != nil {
			return nil, err
		}
	} else {
		if dimgBase, err := c.toDimgBaseDirective(name); err != nil {
			return nil, err
		} else {
			dimg.DimgBase = dimgBase
		}

		if c.RawShell != nil {
			dimg.builder = "shell"
			if shell, err := c.RawShell.toDirective(); err != nil {
				return nil, err
			} else {
				dimg.Shell = shell
			}
		}

		if c.RawDocker != nil {
			if docker, err := c.RawDocker.toDirective(); err != nil {
				return nil, err
			} else {
				dimg.Docker = docker
			}
		}
	}

	if err := c.validateDimgDirective(dimg); err != nil {
		return nil, err
	}

	return
}

func (c *rawDimg) toDimgAsLayersDirective(name string) (dimg *Dimg, err error) {
	var dimgLayers []*Dimg

	var shellDimg *ShellDimg
	var ansible *Ansible
	if c.RawShell != nil {
		shellDimg, err = c.RawShell.toDirective()
		if err != nil {
			return nil, err
		}
	} else if c.RawAnsible != nil {
		ansible, err = c.RawAnsible.toDirective()
		if err != nil {
			return nil, err
		}
	}

	if shellDimg != nil {
		if dimgShellLayers, err := c.toDimgShellLayersByStage(name, shellDimg.BeforeInstall, "beforeInstall"); err != nil {
			return nil, err
		} else {
			dimgLayers = append(dimgLayers, dimgShellLayers...)
		}
	} else if ansible != nil {
		if dimgShellLayers, err := c.toDimgAnsibleLayers(name, ansible.BeforeInstall, "beforeInstall"); err != nil {
			return nil, err
		} else {
			dimgLayers = append(dimgLayers, dimgShellLayers...)
		}
	}

	if dimgWithGit, err := c.toDimgLayerWithGitDirective(name); err != nil {
		return nil, err
	} else if dimgWithGit != nil {
		dimgLayers = append(dimgLayers, dimgWithGit)
	}

	if dimgWithArtifacts, err := c.toDimgLayerWithArtifactsDirective(name, "install", ""); err != nil {
		return nil, err
	} else if dimgWithArtifacts != nil {
		dimgLayers = append(dimgLayers, dimgWithArtifacts)
	}

	if shellDimg != nil {
		if dimgShellLayers, err := c.toDimgShellLayersByStage(name, shellDimg.Install, "install"); err != nil {
			return nil, err
		} else {
			dimgLayers = append(dimgLayers, dimgShellLayers...)
		}
	} else if ansible != nil {
		if dimgShellLayers, err := c.toDimgAnsibleLayers(name, ansible.Install, "install"); err != nil {
			return nil, err
		} else {
			dimgLayers = append(dimgLayers, dimgShellLayers...)
		}
	}

	if dimgWithArtifacts, err := c.toDimgLayerWithArtifactsDirective(name, "", "install"); err != nil {
		return nil, err
	} else if dimgWithArtifacts != nil {
		dimgLayers = append(dimgLayers, dimgWithArtifacts)
	}

	if shellDimg != nil {
		if dimgShellLayers, err := c.toDimgShellLayersByStage(name, shellDimg.BeforeSetup, "beforeSetup"); err != nil {
			return nil, err
		} else {
			dimgLayers = append(dimgLayers, dimgShellLayers...)
		}
	} else if ansible != nil {
		if dimgShellLayers, err := c.toDimgAnsibleLayers(name, ansible.BeforeSetup, "beforeSetup"); err != nil {
			return nil, err
		} else {
			dimgLayers = append(dimgLayers, dimgShellLayers...)
		}
	}

	if dimgWithArtifacts, err := c.toDimgLayerWithArtifactsDirective(name, "setup", ""); err != nil {
		return nil, err
	} else if dimgWithArtifacts != nil {
		dimgLayers = append(dimgLayers, dimgWithArtifacts)
	}

	if shellDimg != nil {
		if dimgShellLayers, err := c.toDimgShellLayersByStage(name, shellDimg.Setup, "setup"); err != nil {
			return nil, err
		} else {
			dimgLayers = append(dimgLayers, dimgShellLayers...)
		}
	} else if ansible != nil {
		if dimgShellLayers, err := c.toDimgAnsibleLayers(name, ansible.Setup, "setup"); err != nil {
			return nil, err
		} else {
			dimgLayers = append(dimgLayers, dimgShellLayers...)
		}
	}

	if dimg, err = c.toMainDimgLayerDirective(name); err != nil {
		return nil, err
	} else {
		dimgLayers = append(dimgLayers, dimg)
	}

	var prevDimgLayer *Dimg
	for _, dimgLayer := range dimgLayers {
		if prevDimgLayer == nil {
			dimgLayer.From = c.From
			dimgLayer.FromCacheVersion = c.FromCacheVersion
		} else {
			dimgLayer.FromDimg = prevDimgLayer
		}
		prevDimgLayer = dimgLayer
	}

	if err = c.validateDimgBaseDirective(prevDimgLayer.DimgBase); err != nil {
		return nil, err
	} else {
		return prevDimgLayer, nil
	}
}

func (c *rawDimg) toDimgShellLayersByStage(name string, commands []string, stage string) (dimgLayers []*Dimg, err error) {
	for ind, command := range commands {
		layerName := fmt.Sprintf("%s-%d", strings.ToLower(stage), ind)
		if name != "" {
			layerName = strings.Join([]string{name, layerName}, "-")
		}

		if dimgLayer, err := c.toDimgLayerDirective(layerName); err != nil {
			return nil, err
		} else {
			dimgLayer.builder = "shell"
			dimgLayer.Shell = c.toShellDimgWithCommandByStage(command, stage)
			dimgLayers = append(dimgLayers, dimgLayer)
		}
	}

	return dimgLayers, nil
}

func (c *rawDimg) toShellDimgWithCommandByStage(command string, stage string) (shellDimg *ShellDimg) {
	shellDimg = &ShellDimg{}
	shellDimg.ShellBase = c.toShellBaseWithCommandByStage(command, stage)
	return
}

func (c *rawDimg) toDimgAnsibleLayers(name string, tasks []*AnsibleTask, stage string) (dimgLayers []*Dimg, err error) {
	for ind, task := range tasks {
		layerName := fmt.Sprintf("%s-%d", strings.ToLower(stage), ind)
		if name != "" {
			layerName = strings.Join([]string{name, layerName}, "-")
		}

		if dimgLayer, err := c.toDimgLayerDirective(layerName); err != nil {
			return nil, err
		} else {
			dimgLayer.builder = "ansible"
			dimgLayer.Ansible = c.toAnsibleWithTaskByStage(task, stage)
			dimgLayers = append(dimgLayers, dimgLayer)
		}
	}

	return dimgLayers, nil
}

func (c *rawDimg) toDimgLayerDirective(layerName string) (dimg *Dimg, err error) {
	dimg = &Dimg{}
	if dimg.DimgBase, err = c.toBaseDimgBaseDirective(layerName); err != nil {
		return nil, err
	}
	return
}

func (c *rawDimg) toDimgLayerWithGitDirective(name string) (dimg *Dimg, err error) {
	if dimgBase, err := c.toLayerWithGitDirective(name); err == nil && dimgBase != nil {
		dimg = &Dimg{}
		dimg.DimgBase = dimgBase
	}
	return
}

func (c *rawDimg) toDimgLayerWithArtifactsDirective(name string, before string, after string) (dimg *Dimg, err error) {
	if dimgBase, err := c.toLayerWithArtifactsDirective(name, before, after); err == nil && dimgBase != nil {
		dimg = &Dimg{}
		dimg.DimgBase = dimgBase
	}
	return
}

func (c *rawDimg) toMainDimgLayerDirective(name string) (mainDimgLayer *Dimg, err error) {
	mainDimgLayer = &Dimg{}
	if mainDimgLayer.DimgBase, err = c.toBaseDimgBaseDirective(name); err != nil {
		return nil, err
	}

	if mainDimgLayer.Import, err = c.layerImportArtifactsByLayer("", "setup"); err != nil {
		return nil, err
	}

	if c.RawDocker != nil {
		if docker, err := c.RawDocker.toDirective(); err != nil {
			return nil, err
		} else {
			mainDimgLayer.Docker = docker
		}
	}

	return
}

func (c *rawDimg) validateDimgDirective(dimg *Dimg) (err error) {
	if err := dimg.validate(); err != nil {
		return err
	}

	return nil
}

func (c *rawDimg) toDimgArtifactAsLayersDirective() (dimgArtifact *DimgArtifact, err error) {
	var dimgArtifactLayers []*DimgArtifact

	var shellArtifact *ShellArtifact
	var ansible *Ansible
	if c.RawShell != nil {
		shellArtifact, err = c.RawShell.toArtifactDirective()
		if err != nil {
			return nil, err
		}
	} else if c.RawAnsible != nil {
		ansible, err = c.RawAnsible.toDirective()
		if err != nil {
			return nil, err
		}
	}

	if shellArtifact != nil {
		if dimgShellLayers, err := c.toDimgArtifactShellLayers(shellArtifact.BeforeInstall, "beforeInstall"); err != nil {
			return nil, err
		} else {
			dimgArtifactLayers = append(dimgArtifactLayers, dimgShellLayers...)
		}
	} else if ansible != nil {
		if dimgShellLayers, err := c.toDimgArtifactAnsibleLayers(ansible.BeforeInstall, "beforeInstall"); err != nil {
			return nil, err
		} else {
			dimgArtifactLayers = append(dimgArtifactLayers, dimgShellLayers...)
		}
	}

	if dimgWithArtifacts, err := c.toDimgArtifactLayerWithArtifactsDirective("install", ""); err != nil {
		return nil, err
	} else if dimgWithArtifacts != nil {
		dimgArtifactLayers = append(dimgArtifactLayers, dimgWithArtifacts)
	}

	if dimgArtifactWithGit, err := c.toDimgArtifactLayerWithGitDirective(); err != nil {
		return nil, err
	} else if dimgArtifactWithGit != nil {
		dimgArtifactLayers = append(dimgArtifactLayers, dimgArtifactWithGit)
	}

	if shellArtifact != nil {
		if dimgShellLayers, err := c.toDimgArtifactShellLayers(shellArtifact.Install, "install"); err != nil {
			return nil, err
		} else {
			dimgArtifactLayers = append(dimgArtifactLayers, dimgShellLayers...)
		}
	} else if ansible != nil {
		if dimgShellLayers, err := c.toDimgArtifactAnsibleLayers(ansible.Install, "install"); err != nil {
			return nil, err
		} else {
			dimgArtifactLayers = append(dimgArtifactLayers, dimgShellLayers...)
		}
	}

	if dimgWithArtifacts, err := c.toDimgArtifactLayerWithArtifactsDirective("", "install"); err != nil {
		return nil, err
	} else if dimgWithArtifacts != nil {
		dimgArtifactLayers = append(dimgArtifactLayers, dimgWithArtifacts)
	}

	if shellArtifact != nil {
		if dimgShellLayers, err := c.toDimgArtifactShellLayers(shellArtifact.BeforeSetup, "beforeSetup"); err != nil {
			return nil, err
		} else {
			dimgArtifactLayers = append(dimgArtifactLayers, dimgShellLayers...)
		}
	} else if ansible != nil {
		if dimgShellLayers, err := c.toDimgArtifactAnsibleLayers(ansible.BeforeSetup, "beforeSetup"); err != nil {
			return nil, err
		} else {
			dimgArtifactLayers = append(dimgArtifactLayers, dimgShellLayers...)
		}
	}

	if dimgWithArtifacts, err := c.toDimgArtifactLayerWithArtifactsDirective("setup", ""); err != nil {
		return nil, err
	} else if dimgWithArtifacts != nil {
		dimgArtifactLayers = append(dimgArtifactLayers, dimgWithArtifacts)
	}

	if shellArtifact != nil {
		if dimgShellLayers, err := c.toDimgArtifactShellLayers(shellArtifact.Setup, "setup"); err != nil {
			return nil, err
		} else {
			dimgArtifactLayers = append(dimgArtifactLayers, dimgShellLayers...)
		}
	} else if ansible != nil {
		if dimgShellLayers, err := c.toDimgArtifactAnsibleLayers(ansible.Setup, "setup"); err != nil {
			return nil, err
		} else {
			dimgArtifactLayers = append(dimgArtifactLayers, dimgShellLayers...)
		}
	}

	if dimgWithArtifacts, err := c.toDimgArtifactLayerWithArtifactsDirective("", "setup"); err != nil {
		return nil, err
	} else if dimgWithArtifacts != nil {
		dimgArtifactLayers = append(dimgArtifactLayers, dimgWithArtifacts)
	}

	if shellArtifact != nil {
		if dimgShellLayers, err := c.toDimgArtifactShellLayers(shellArtifact.BuildArtifact, "buildArtifact"); err != nil {
			return nil, err
		} else {
			dimgArtifactLayers = append(dimgArtifactLayers, dimgShellLayers...)
		}
	} else if ansible != nil {
		if dimgShellLayers, err := c.toDimgArtifactAnsibleLayers(ansible.BuildArtifact, "buildArtifact"); err != nil {
			return nil, err
		} else {
			dimgArtifactLayers = append(dimgArtifactLayers, dimgShellLayers...)
		}
	}

	if dimgArtifact, err = c.toMainDimgArtifactLayerDirective(); err != nil {
		return nil, err
	} else {
		dimgArtifactLayers = append(dimgArtifactLayers, dimgArtifact)
	}

	var prevDimgLayer *DimgArtifact
	for _, dimgArtifactLayer := range dimgArtifactLayers {
		if prevDimgLayer == nil {
			dimgArtifactLayer.From = c.From
			dimgArtifactLayer.FromCacheVersion = c.FromCacheVersion
		} else {
			dimgArtifactLayer.FromDimgArtifact = prevDimgLayer
		}
		prevDimgLayer = dimgArtifactLayer
	}

	if err = c.validateDimgBaseDirective(prevDimgLayer.DimgBase); err != nil {
		return nil, err
	} else {
		return prevDimgLayer, nil
	}
}

func (c *rawDimg) toDimgArtifactLayerDirective(layerName string) (dimgArtifact *DimgArtifact, err error) {
	dimgArtifact = &DimgArtifact{}
	if dimgArtifact.DimgBase, err = c.toBaseDimgBaseDirective(layerName); err != nil {
		return nil, err
	}
	return
}

func (c *rawDimg) toDimgArtifactLayerWithGitDirective() (dimgArtifact *DimgArtifact, err error) {
	if dimgBase, err := c.toLayerWithGitDirective(c.Artifact); err == nil && dimgBase != nil {
		dimgArtifact = &DimgArtifact{}
		dimgArtifact.DimgBase = dimgBase
	}
	return
}

func (c *rawDimg) toDimgArtifactLayerWithArtifactsDirective(before string, after string) (dimgArtifact *DimgArtifact, err error) {
	if dimgBase, err := c.toLayerWithArtifactsDirective(c.Artifact, before, after); err == nil && dimgBase != nil {
		dimgArtifact = &DimgArtifact{}
		dimgArtifact.DimgBase = dimgBase
	}
	return
}

func (c *rawDimg) toLayerWithGitDirective(name string) (dimgBase *DimgBase, err error) {
	if len(c.RawGit) != 0 {
		layerName := "git"
		if name != "" {
			layerName = strings.Join([]string{name, layerName}, "-")
		}

		if dimgBase, err = c.toBaseDimgBaseDirective(layerName); err != nil {
			return nil, err
		}

		dimgBase.Git = &GitManager{}
		for _, git := range c.RawGit {
			if git.gitType() == "local" {
				if gitLocal, err := git.toGitLocalDirective(); err != nil {
					return nil, err
				} else {
					dimgBase.Git.Local = append(dimgBase.Git.Local, gitLocal)
				}
			} else {
				if gitRemote, err := git.toGitRemoteDirective(); err != nil {
					return nil, err
				} else {
					dimgBase.Git.Remote = append(dimgBase.Git.Remote, gitRemote)
				}
			}
		}
	}

	return
}

func (c *rawDimg) toLayerWithArtifactsDirective(name string, before string, after string) (dimgBase *DimgBase, err error) {
	if importArtifacts, err := c.layerImportArtifactsByLayer(before, after); err != nil {
		return nil, err
	} else {
		if len(importArtifacts) != 0 {
			var layerName string
			if before != "" {
				layerName = fmt.Sprintf("before-%s-artifacts", before)
			} else {
				layerName = fmt.Sprintf("after-%s-artifacts", after)
			}

			if name != "" {
				layerName = strings.Join([]string{name, layerName}, "-")
			}

			if dimgBase, err = c.toBaseDimgBaseDirective(layerName); err != nil {
				return nil, err
			}
			dimgBase.Import = importArtifacts
		} else {
			return nil, nil
		}
	}

	return
}

func (c *rawDimg) toMainDimgArtifactLayerDirective() (mainDimgArtifactLayer *DimgArtifact, err error) {
	mainDimgArtifactLayer = &DimgArtifact{}
	if mainDimgArtifactLayer.DimgBase, err = c.toBaseDimgBaseDirective(c.Artifact); err != nil {
		return nil, err
	}
	return mainDimgArtifactLayer, nil
}

func (c *rawDimg) layerImportArtifactsByLayer(before string, after string) (artifactImports []*ArtifactImport, err error) {
	for _, importArtifact := range c.RawImport {
		var condition bool
		if before != "" {
			condition = importArtifact.Before == before
		} else {
			condition = importArtifact.After == after
		}

		if !condition {
			continue
		}

		if importArtifactDirective, err := importArtifact.toDirective(); err != nil {
			return nil, err
		} else {
			artifactImports = append(artifactImports, importArtifactDirective)
		}
	}

	return
}

func (c *rawDimg) toDimgArtifactShellLayers(commands []string, stage string) (dimgArtifactLayers []*DimgArtifact, err error) {
	for ind, command := range commands {
		layerName := fmt.Sprintf("%s-%s-%d", c.Artifact, strings.ToLower(stage), ind)
		if dimgArtifactLayer, err := c.toDimgArtifactLayerDirective(layerName); err != nil {
			return nil, err
		} else {
			dimgArtifactLayer.builder = "shell"
			dimgArtifactLayer.Shell = c.toShellArtifactWithCommandByStage(command, stage)
			dimgArtifactLayers = append(dimgArtifactLayers, dimgArtifactLayer)
		}
	}

	return dimgArtifactLayers, nil
}

func (c *rawDimg) toShellArtifactWithCommandByStage(command string, stage string) (shellArtifact *ShellArtifact) {
	shellArtifact = &ShellArtifact{}
	if stage == "buildArtifact" {
		shellArtifact.BuildArtifact = []string{command}
	} else {
		shellArtifact.ShellDimg = c.toShellDimgWithCommandByStage(command, stage)
	}
	return
}

func (c *rawDimg) toShellBaseWithCommandByStage(command string, stage string) (shellBase *ShellBase) {
	shellBase = &ShellBase{}
	switch stage {
	case "beforeInstall":
		shellBase.BeforeInstall = []string{command}
	case "install":
		shellBase.Install = []string{command}
	case "beforeSetup":
		shellBase.BeforeSetup = []string{command}
	case "setup":
		shellBase.Setup = []string{command}
	}
	shellBase.raw = c.RawShell
	return
}

func (c *rawDimg) toDimgArtifactAnsibleLayers(tasks []*AnsibleTask, stage string) (dimgLayers []*DimgArtifact, err error) {
	for ind, task := range tasks {
		layerName := fmt.Sprintf("%s-%s-%d", c.Artifact, strings.ToLower(stage), ind)
		if dimgLayer, err := c.toDimgArtifactLayerDirective(layerName); err != nil {
			return nil, err
		} else {
			dimgLayer.builder = "ansible"
			dimgLayer.Ansible = c.toAnsibleWithTaskByStage(task, stage)
			dimgLayers = append(dimgLayers, dimgLayer)
		}
	}

	return dimgLayers, nil
}

func (c *rawDimg) toAnsibleWithTaskByStage(task *AnsibleTask, stage string) (ansible *Ansible) {
	ansible = &Ansible{}
	switch stage {
	case "beforeInstall":
		ansible.BeforeInstall = []*AnsibleTask{task}
	case "install":
		ansible.Install = []*AnsibleTask{task}
	case "beforeSetup":
		ansible.BeforeSetup = []*AnsibleTask{task}
	case "setup":
		ansible.Setup = []*AnsibleTask{task}
	case "buildArtifact":
		ansible.BuildArtifact = []*AnsibleTask{task}
	}
	ansible.raw = c.RawAnsible
	return
}

func (c *rawDimg) validateArtifactDimgDirective(dimgArtifact *DimgArtifact) (err error) {
	if c.RawDocker != nil {
		return newDetailedConfigError("`docker` section is not supported for artifact!", nil, c.doc)
	}

	if err := dimgArtifact.validate(); err != nil {
		return err
	}

	return nil
}

func (c *rawDimg) toDimgBaseDirective(name string) (dimgBase *DimgBase, err error) {
	if dimgBase, err = c.toBaseDimgBaseDirective(name); err != nil {
		return nil, err
	}

	dimgBase.From = c.From
	dimgBase.FromCacheVersion = c.FromCacheVersion

	dimgBase.Git = &GitManager{}
	for _, git := range c.RawGit {
		if git.gitType() == "local" {
			if gitLocal, err := git.toGitLocalDirective(); err != nil {
				return nil, err
			} else {
				dimgBase.Git.Local = append(dimgBase.Git.Local, gitLocal)
			}
		} else {
			if gitRemote, err := git.toGitRemoteDirective(); err != nil {
				return nil, err
			} else {
				dimgBase.Git.Remote = append(dimgBase.Git.Remote, gitRemote)
			}
		}
	}

	if c.RawAnsible != nil {
		dimgBase.builder = "ansible"
		if ansible, err := c.RawAnsible.toDirective(); err != nil {
			return nil, err
		} else {
			dimgBase.Ansible = ansible
		}
	}

	for _, importArtifact := range c.RawImport {
		if importArtifactDirective, err := importArtifact.toDirective(); err != nil {
			return nil, err
		} else {
			dimgBase.Import = append(dimgBase.Import, importArtifactDirective)
		}
	}

	if err := c.validateDimgBaseDirective(dimgBase); err != nil {
		return nil, err
	}

	return dimgBase, nil
}

func (c *rawDimg) validateDimgBaseDirective(dimgBase *DimgBase) (err error) {
	if err := dimgBase.validate(); err != nil {
		return err
	}

	return nil
}

func (c *rawDimg) toBaseDimgBaseDirective(name string) (dimgBase *DimgBase, err error) {
	dimgBase = &DimgBase{}
	dimgBase.Name = name
	dimgBase.builder = "none"

	for _, mount := range c.RawMount {
		if dimgMount, err := mount.toDirective(); err != nil {
			return nil, err
		} else {
			dimgBase.Mount = append(dimgBase.Mount, dimgMount)
		}
	}

	dimgBase.raw = c

	return dimgBase, nil
}
