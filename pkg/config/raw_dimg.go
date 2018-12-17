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

func (c *rawDimg) setAndValidateDimg() error {
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
			return newDetailedConfigError(fmt.Sprintf("invalid dimg name `%v`!", t), nil, c.doc)
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

	if err := c.setAndValidateDimg(); err != nil {
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
		return newDetailedConfigError("unknown doc type: one and only one of `dimg: NAME` or `artifact: NAME` non-empty name required!", nil, c.doc)
	} else if !(isDimg || isArtifact) {
		return newDetailedConfigError("unknown doc type: one of `dimg: NAME` or `artifact: NAME` non-empty name required!", nil, c.doc)
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
	dimgBaseLayers, err := c.toDimgBaseLayersDirectives(name)
	if err != nil {
		return nil, err
	}

	var layers []*Dimg
	for _, dimgBaseLayer := range dimgBaseLayers {
		layer := &Dimg{}
		layer.DimgBase = dimgBaseLayer
		layers = append(layers, layer)
	}

	if dimg, err = c.toDimgTopLayerDirective(name); err != nil {
		return nil, err
	} else {
		layers = append(layers, dimg)
	}

	var prevDimgLayer *Dimg
	for _, dimgLayer := range layers {
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

func (c *rawDimg) toDimgBaseLayersDirectives(name string) (layers []*DimgBase, err error) {
	var shell *Shell
	var ansible *Ansible
	if c.RawShell != nil {
		shell, err = c.RawShell.toDirective()
		if err != nil {
			return nil, err
		}
	} else if c.RawAnsible != nil {
		ansible, err = c.RawAnsible.toDirective()
		if err != nil {
			return nil, err
		}
	}

	if shell != nil {
		if beforeInstallShellLayers, err := c.toDimgBaseShellLayersDirectivesByStage(name, shell.BeforeInstall, "beforeInstall"); err != nil {
			return nil, err
		} else {
			layers = append(layers, beforeInstallShellLayers...)
		}
	} else if ansible != nil {
		if beforeInstallAnsibleLayers, err := c.toDimgBaseAnsibleLayersDirectivesByStage(name, ansible.BeforeInstall, "beforeInstall"); err != nil {
			return nil, err
		} else {
			layers = append(layers, beforeInstallAnsibleLayers...)
		}
	}

	if gitLayer, err := c.toDimgBaseGitLayerDirective(name); err != nil {
		return nil, err
	} else if gitLayer != nil {
		layers = append(layers, gitLayer)
	}

	if importsLayers, err := c.toDimgBaseImportsLayerDirectiveByBeforeAndAfter(name, "install", ""); err != nil {
		return nil, err
	} else if importsLayers != nil {
		layers = append(layers, importsLayers)
	}

	if shell != nil {
		if installShellLayers, err := c.toDimgBaseShellLayersDirectivesByStage(name, shell.Install, "install"); err != nil {
			return nil, err
		} else {
			layers = append(layers, installShellLayers...)
		}
	} else if ansible != nil {
		if installAnsibleLayers, err := c.toDimgBaseAnsibleLayersDirectivesByStage(name, ansible.Install, "install"); err != nil {
			return nil, err
		} else {
			layers = append(layers, installAnsibleLayers...)
		}
	}

	if importsLayer, err := c.toDimgBaseImportsLayerDirectiveByBeforeAndAfter(name, "", "install"); err != nil {
		return nil, err
	} else if importsLayer != nil {
		layers = append(layers, importsLayer)
	}

	if shell != nil {
		if beforeSetupShellLayers, err := c.toDimgBaseShellLayersDirectivesByStage(name, shell.BeforeSetup, "beforeSetup"); err != nil {
			return nil, err
		} else {
			layers = append(layers, beforeSetupShellLayers...)
		}
	} else if ansible != nil {
		if beforeSetupAnsibleLayers, err := c.toDimgBaseAnsibleLayersDirectivesByStage(name, ansible.BeforeSetup, "beforeSetup"); err != nil {
			return nil, err
		} else {
			layers = append(layers, beforeSetupAnsibleLayers...)
		}
	}

	if importsLayer, err := c.toDimgBaseImportsLayerDirectiveByBeforeAndAfter(name, "setup", ""); err != nil {
		return nil, err
	} else if importsLayer != nil {
		layers = append(layers, importsLayer)
	}

	if shell != nil {
		if setupShellLayers, err := c.toDimgBaseShellLayersDirectivesByStage(name, shell.Setup, "setup"); err != nil {
			return nil, err
		} else {
			layers = append(layers, setupShellLayers...)
		}
	} else if ansible != nil {
		if setupAnsibleLayers, err := c.toDimgBaseAnsibleLayersDirectivesByStage(name, ansible.Setup, "setup"); err != nil {
			return nil, err
		} else {
			layers = append(layers, setupAnsibleLayers...)
		}
	}

	return layers, nil
}

func (c *rawDimg) toDimgBaseShellLayersDirectivesByStage(name string, commands []string, stage string) (dimgLayers []*DimgBase, err error) {
	for ind, command := range commands {
		layerName := fmt.Sprintf("%s-%d", strings.ToLower(stage), ind)
		if name != "" {
			layerName = strings.Join([]string{name, layerName}, "-")
		}

		if dimgBaseLayer, err := c.toBaseDimgBaseDirective(layerName); err != nil {
			return nil, err
		} else {
			dimgBaseLayer.Shell = c.toShellDirectiveByCommandAndStage(command, stage)
			dimgLayers = append(dimgLayers, dimgBaseLayer)
		}
	}

	return dimgLayers, nil
}

func (c *rawDimg) toDimgBaseAnsibleLayersDirectivesByStage(name string, tasks []*AnsibleTask, stage string) (layers []*DimgBase, err error) {
	for ind, task := range tasks {
		layerName := fmt.Sprintf("%s-%d", strings.ToLower(stage), ind)
		if name != "" {
			layerName = strings.Join([]string{name, layerName}, "-")
		}

		if layer, err := c.toBaseDimgBaseDirective(layerName); err != nil {
			return nil, err
		} else {
			layer.Ansible = c.toAnsibleWithTaskByStage(task, stage)
			layers = append(layers, layer)
		}
	}

	return layers, nil
}

func (c *rawDimg) toDimgLayerDirective(layerName string) (dimg *Dimg, err error) {
	dimg = &Dimg{}
	if dimg.DimgBase, err = c.toBaseDimgBaseDirective(layerName); err != nil {
		return nil, err
	}
	return
}

func (c *rawDimg) toDimgTopLayerDirective(name string) (mainDimgLayer *Dimg, err error) {
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

func (c *rawDimg) toDimgArtifactAsLayersDirective() (dimgArtifactLayer *DimgArtifact, err error) {
	dimgBaseLayers, err := c.toDimgBaseLayersDirectives(c.Artifact)
	if err != nil {
		return nil, err
	}

	var layers []*DimgArtifact
	for _, dimgBaseLayer := range dimgBaseLayers {
		layer := &DimgArtifact{}
		layer.DimgBase = dimgBaseLayer
		layers = append(layers, layer)
	}

	if dimgArtifactLayer, err = c.toDimgArtifactTopLayerDirective(); err != nil {
		return nil, err
	} else {
		layers = append(layers, dimgArtifactLayer)
	}

	var prevDimgLayer *DimgArtifact
	for _, layer := range layers {
		if prevDimgLayer == nil {
			layer.From = c.From
			layer.FromCacheVersion = c.FromCacheVersion
		} else {
			layer.FromDimgArtifact = prevDimgLayer
		}

		prevDimgLayer = layer
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
	if dimgBase, err := c.toDimgBaseGitLayerDirective(c.Artifact); err == nil && dimgBase != nil {
		dimgArtifact = &DimgArtifact{}
		dimgArtifact.DimgBase = dimgBase
	}
	return
}

func (c *rawDimg) toDimgArtifactLayerWithArtifactsDirective(before string, after string) (dimgArtifact *DimgArtifact, err error) {
	if dimgBase, err := c.toDimgBaseImportsLayerDirectiveByBeforeAndAfter(c.Artifact, before, after); err == nil && dimgBase != nil {
		dimgArtifact = &DimgArtifact{}
		dimgArtifact.DimgBase = dimgBase
	}
	return
}

func (c *rawDimg) toDimgBaseGitLayerDirective(name string) (dimgBase *DimgBase, err error) {
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

func (c *rawDimg) toDimgBaseImportsLayerDirectiveByBeforeAndAfter(name string, before string, after string) (dimgBase *DimgBase, err error) {
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

func (c *rawDimg) toDimgArtifactTopLayerDirective() (mainDimgArtifactLayer *DimgArtifact, err error) {
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

func (c *rawDimg) toShellDirectiveByCommandAndStage(command string, stage string) (shell *Shell) {
	shell = &Shell{}
	switch stage {
	case "beforeInstall":
		shell.BeforeInstall = []string{command}
	case "install":
		shell.Install = []string{command}
	case "beforeSetup":
		shell.BeforeSetup = []string{command}
	case "setup":
		shell.Setup = []string{command}
	}

	shell.raw = c.RawShell

	return
}

func (c *rawDimg) toDimgArtifactAnsibleLayers(tasks []*AnsibleTask, stage string) (dimgLayers []*DimgArtifact, err error) {
	for ind, task := range tasks {
		layerName := fmt.Sprintf("%s-%s-%d", c.Artifact, strings.ToLower(stage), ind)
		if dimgLayer, err := c.toDimgArtifactLayerDirective(layerName); err != nil {
			return nil, err
		} else {
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

	if c.RawShell != nil {
		if shell, err := c.RawShell.toDirective(); err != nil {
			return nil, err
		} else {
			dimgBase.Shell = shell
		}
	}

	if c.RawAnsible != nil {
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

	for _, mount := range c.RawMount {
		if dimgMount, err := mount.toDirective(); err != nil {
			return nil, err
		} else {
			dimgBase.Mount = append(dimgBase.Mount, dimgMount)
		}
	}

	dimgBase.Git = &GitManager{}

	dimgBase.raw = c

	return dimgBase, nil
}
