package config

import (
	"fmt"
	"strings"
)

type rawStapelImage struct {
	Images           []string     `yaml:"-"`
	Artifact         string       `yaml:"artifact,omitempty"`
	From             string       `yaml:"from,omitempty"`
	FromLatest       bool         `yaml:"fromLatest,omitempty"`
	FromCacheVersion string       `yaml:"fromCacheVersion,omitempty"`
	FromImage        string       `yaml:"fromImage,omitempty"`
	FromArtifact     string       `yaml:"fromArtifact,omitempty"`
	RawGit           []*rawGit    `yaml:"git,omitempty"`
	RawShell         *rawShell    `yaml:"shell,omitempty"`
	RawAnsible       *rawAnsible  `yaml:"ansible,omitempty"`
	RawMount         []*rawMount  `yaml:"mount,omitempty"`
	RawDocker        *rawDocker   `yaml:"docker,omitempty"`
	RawImport        []*rawImport `yaml:"import,omitempty"`
	AsLayers         bool         `yaml:"asLayers,omitempty"`

	doc                *doc `yaml:"-"` // parent
	DisableDeterminism bool `yaml:"-"` // parser option

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *rawStapelImage) setAndValidateStapelImage() error {
	value, ok := c.UnsupportedAttributes["image"]
	if ok {
		delete(c.UnsupportedAttributes, "image")

		switch t := value.(type) {
		case []interface{}:
			if images, err := InterfaceToStringArray(value, nil, c.doc); err != nil {
				return err
			} else {
				c.Images = images
			}
		case string:
			c.Images = []string{value.(string)}
		case nil:
			c.Images = []string{""}
		default:
			return newDetailedConfigError(fmt.Sprintf("invalid image name `%v`!", t), nil, c.doc)
		}
	}

	return nil
}

func (c *rawStapelImage) UnmarshalYAML(unmarshal func(interface{}) error) error {
	parentStack.Push(c)
	type plain rawStapelImage
	err := unmarshal((*plain)(c))
	parentStack.Pop()
	if err != nil {
		return err
	}

	if err := c.setAndValidateStapelImage(); err != nil {
		return err
	}

	if err := checkOverflow(c.UnsupportedAttributes, nil, c.doc); err != nil {
		return err
	}

	if err := c.validateStapelImageType(); err != nil {
		return err
	}

	return nil
}

func (c *rawStapelImage) validateStapelImageType() error {
	isImage := len(c.Images) != 0
	isArtifact := c.Artifact != ""

	if isImage && isArtifact {
		return newDetailedConfigError("unknown doc type: one and only one of `image: NAME` or `artifact: NAME` non-empty name required!", nil, c.doc)
	} else if !(isImage || isArtifact) {
		return newDetailedConfigError("unknown doc type: one of `image: NAME` or `artifact: NAME` non-empty name required!", nil, c.doc)
	}

	return nil
}

func (c *rawStapelImage) stapelImageType() string {
	if len(c.Images) != 0 {
		return "images"
	} else if c.Artifact != "" {
		return "artifact"
	}

	return ""
}

func (c *rawStapelImage) toStapelImageDirectives() (images []*StapelImage, err error) {
	for _, imageName := range c.Images {
		if imageImages, err := c.toStapelImageDirectiveGroup(imageName); err != nil {
			return nil, err
		} else {
			images = append(images, imageImages...)
		}
	}

	return images, nil
}

func (c *rawStapelImage) toStapelImageArtifactDirectives() (imageArtifacts []*StapelImageArtifact, err error) {
	if c.AsLayers {
		if imageArtifactLayers, err := c.toStapelImageArtifactAsLayersDirective(); err != nil {
			return nil, err
		} else {
			imageArtifacts = append(imageArtifacts, imageArtifactLayers...)
		}
	} else {
		imageArtifact := &StapelImageArtifact{}
		if imageArtifact.StapelImageBase, err = c.toStapelImageBaseDirective(c.Artifact); err != nil {
			return nil, err
		}

		imageArtifacts = append(imageArtifacts, imageArtifact)
	}

	for _, imageArtifact := range imageArtifacts {
		if err := c.validateStapelImageArtifactDirective(imageArtifact); err != nil {
			return nil, err
		}
	}

	return imageArtifacts, nil
}

func (c *rawStapelImage) toStapelImageDirectiveGroup(name string) (images []*StapelImage, err error) {
	image := &StapelImage{}

	if c.AsLayers {
		if imageLayers, err := c.toImageAsLayersDirective(name); err != nil {
			return nil, err
		} else {
			images = append(images, imageLayers...)
		}
	} else {
		if imageBase, err := c.toStapelImageBaseDirective(name); err != nil {
			return nil, err
		} else {
			image.StapelImageBase = imageBase
		}

		if c.RawDocker != nil {
			if docker, err := c.RawDocker.toDirective(); err != nil {
				return nil, err
			} else {
				image.Docker = docker
			}
		}

		images = append(images, image)
	}

	for _, image := range images {
		if err := c.validateStapelImageDirective(image); err != nil {
			return nil, err
		}
	}

	return
}

func (c *rawStapelImage) toImageAsLayersDirective(name string) (imageLayers []*StapelImage, err error) {
	image := &StapelImage{}

	imageBaseLayers, err := c.toStapelImageBaseLayersDirectives(name)
	if err != nil {
		return nil, err
	}

	for _, imageBaseLayer := range imageBaseLayers {
		layer := &StapelImage{}
		layer.StapelImageBase = imageBaseLayer
		imageLayers = append(imageLayers, layer)
	}

	if image, err = c.toStapelImageTopLayerDirective(name); err != nil {
		return nil, err
	} else {
		imageLayers = append(imageLayers, image)
	}

	var prevImageLayer *StapelImage
	for _, layer := range imageLayers {
		if prevImageLayer == nil {
			layer.From = c.From
			layer.FromImageName = c.FromImage
			layer.FromArtifactName = c.FromArtifact
			layer.FromLatest = c.FromLatest
			layer.FromCacheVersion = c.FromCacheVersion
		} else {
			layer.FromImageName = prevImageLayer.Name
		}
		prevImageLayer = layer
	}

	if err = c.validateStapelImageBaseDirective(prevImageLayer.StapelImageBase); err != nil {
		return nil, err
	} else {
		return imageLayers, nil
	}
}

func (c *rawStapelImage) toStapelImageBaseLayersDirectives(name string) (imageLayers []*StapelImageBase, err error) {
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
		if beforeInstallShellLayers, err := c.toStapelImageBaseShellLayersDirectivesByStage(name, shell.BeforeInstall, "beforeInstall"); err != nil {
			return nil, err
		} else {
			imageLayers = append(imageLayers, beforeInstallShellLayers...)
		}
	} else if ansible != nil {
		if beforeInstallAnsibleLayers, err := c.toStapelImageBaseAnsibleLayersDirectivesByStage(name, ansible.BeforeInstall, "beforeInstall"); err != nil {
			return nil, err
		} else {
			imageLayers = append(imageLayers, beforeInstallAnsibleLayers...)
		}
	}

	if gitLayer, err := c.toStapelImageBaseGitLayerDirective(name); err != nil {
		return nil, err
	} else if gitLayer != nil {
		imageLayers = append(imageLayers, gitLayer)
	}

	if importsLayers, err := c.toStapelImageBaseImportsLayerDirectiveByBeforeAndAfter(name, "install", ""); err != nil {
		return nil, err
	} else if importsLayers != nil {
		imageLayers = append(imageLayers, importsLayers)
	}

	if shell != nil {
		if installShellLayers, err := c.toStapelImageBaseShellLayersDirectivesByStage(name, shell.Install, "install"); err != nil {
			return nil, err
		} else {
			imageLayers = append(imageLayers, installShellLayers...)
		}
	} else if ansible != nil {
		if installAnsibleLayers, err := c.toStapelImageBaseAnsibleLayersDirectivesByStage(name, ansible.Install, "install"); err != nil {
			return nil, err
		} else {
			imageLayers = append(imageLayers, installAnsibleLayers...)
		}
	}

	if importsLayer, err := c.toStapelImageBaseImportsLayerDirectiveByBeforeAndAfter(name, "", "install"); err != nil {
		return nil, err
	} else if importsLayer != nil {
		imageLayers = append(imageLayers, importsLayer)
	}

	if shell != nil {
		if beforeSetupShellLayers, err := c.toStapelImageBaseShellLayersDirectivesByStage(name, shell.BeforeSetup, "beforeSetup"); err != nil {
			return nil, err
		} else {
			imageLayers = append(imageLayers, beforeSetupShellLayers...)
		}
	} else if ansible != nil {
		if beforeSetupAnsibleLayers, err := c.toStapelImageBaseAnsibleLayersDirectivesByStage(name, ansible.BeforeSetup, "beforeSetup"); err != nil {
			return nil, err
		} else {
			imageLayers = append(imageLayers, beforeSetupAnsibleLayers...)
		}
	}

	if importsLayer, err := c.toStapelImageBaseImportsLayerDirectiveByBeforeAndAfter(name, "setup", ""); err != nil {
		return nil, err
	} else if importsLayer != nil {
		imageLayers = append(imageLayers, importsLayer)
	}

	if shell != nil {
		if setupShellLayers, err := c.toStapelImageBaseShellLayersDirectivesByStage(name, shell.Setup, "setup"); err != nil {
			return nil, err
		} else {
			imageLayers = append(imageLayers, setupShellLayers...)
		}
	} else if ansible != nil {
		if setupAnsibleLayers, err := c.toStapelImageBaseAnsibleLayersDirectivesByStage(name, ansible.Setup, "setup"); err != nil {
			return nil, err
		} else {
			imageLayers = append(imageLayers, setupAnsibleLayers...)
		}
	}

	return imageLayers, nil
}

func (c *rawStapelImage) toStapelImageBaseShellLayersDirectivesByStage(name string, commands []string, stage string) (imageLayers []*StapelImageBase, err error) {
	for ind, command := range commands {
		layerName := fmt.Sprintf("%s-%d", strings.ToLower(stage), ind)
		if name != "" {
			layerName = strings.Join([]string{name, layerName}, "-")
		}

		if imageBaseLayer, err := c.toBaseStapelImageBaseDirective(layerName); err != nil {
			return nil, err
		} else {
			imageBaseLayer.Shell = c.toShellDirectiveByCommandAndStage(command, stage)
			imageLayers = append(imageLayers, imageBaseLayer)
		}
	}

	return imageLayers, nil
}

func (c *rawStapelImage) toStapelImageBaseAnsibleLayersDirectivesByStage(name string, tasks []*AnsibleTask, stage string) (imageLayers []*StapelImageBase, err error) {
	for ind, task := range tasks {
		layerName := fmt.Sprintf("%s-%d", strings.ToLower(stage), ind)
		if name != "" {
			layerName = strings.Join([]string{name, layerName}, "-")
		}

		if layer, err := c.toBaseStapelImageBaseDirective(layerName); err != nil {
			return nil, err
		} else {
			layer.Ansible = c.toAnsibleWithTaskByStage(task, stage)
			imageLayers = append(imageLayers, layer)
		}
	}

	return imageLayers, nil
}

func (c *rawStapelImage) toStapelImageTopLayerDirective(name string) (mainImageLayer *StapelImage, err error) {
	mainImageLayer = &StapelImage{}
	if mainImageLayer.StapelImageBase, err = c.toBaseStapelImageBaseDirective(name); err != nil {
		return nil, err
	}

	if mainImageLayer.Import, err = c.layerStapelImportArtifactsByLayer("", "setup"); err != nil {
		return nil, err
	}

	if c.RawDocker != nil {
		if docker, err := c.RawDocker.toDirective(); err != nil {
			return nil, err
		} else {
			mainImageLayer.Docker = docker
		}
	}

	return
}

func (c *rawStapelImage) validateStapelImageDirective(image *StapelImage) (err error) {
	if err := image.validate(); err != nil {
		return err
	}

	return nil
}

func (c *rawStapelImage) toStapelImageArtifactAsLayersDirective() (imageArtifactLayers []*StapelImageArtifact, err error) {
	imageArtifactLayer := &StapelImageArtifact{}

	imageBaseLayers, err := c.toStapelImageBaseLayersDirectives(c.Artifact)
	if err != nil {
		return nil, err
	}

	for _, imageBaseLayer := range imageBaseLayers {
		layer := &StapelImageArtifact{}
		layer.StapelImageBase = imageBaseLayer
		imageArtifactLayers = append(imageArtifactLayers, layer)
	}

	if imageArtifactLayer, err = c.toStapelImageArtifactTopLayerDirective(); err != nil {
		return nil, err
	} else {
		imageArtifactLayers = append(imageArtifactLayers, imageArtifactLayer)
	}

	var prevImageLayer *StapelImageArtifact
	for _, layer := range imageArtifactLayers {
		if prevImageLayer == nil {
			layer.From = c.From
			layer.FromImageName = c.FromImage
			layer.FromArtifactName = c.FromArtifact
			layer.FromLatest = c.FromLatest
			layer.FromCacheVersion = c.FromCacheVersion
		} else {
			layer.FromArtifactName = prevImageLayer.Name
		}

		prevImageLayer = layer
	}

	if err = c.validateStapelImageBaseDirective(prevImageLayer.StapelImageBase); err != nil {
		return nil, err
	} else {
		return imageArtifactLayers, nil
	}
}

func (c *rawStapelImage) toStapelImageArtifactLayerDirective(layerName string) (imageArtifact *StapelImageArtifact, err error) {
	imageArtifact = &StapelImageArtifact{}
	if imageArtifact.StapelImageBase, err = c.toBaseStapelImageBaseDirective(layerName); err != nil {
		return nil, err
	}
	return
}

func (c *rawStapelImage) toImageArtifactLayerWithGitDirective() (imageArtifact *StapelImageArtifact, err error) {
	if imageBase, err := c.toStapelImageBaseGitLayerDirective(c.Artifact); err == nil && imageBase != nil {
		imageArtifact = &StapelImageArtifact{}
		imageArtifact.StapelImageBase = imageBase
	}
	return
}

func (c *rawStapelImage) toStapelImageArtifactLayerWithArtifactsDirective(before string, after string) (imageArtifact *StapelImageArtifact, err error) {
	if imageBase, err := c.toStapelImageBaseImportsLayerDirectiveByBeforeAndAfter(c.Artifact, before, after); err == nil && imageBase != nil {
		imageArtifact = &StapelImageArtifact{}
		imageArtifact.StapelImageBase = imageBase
	}
	return
}

func (c *rawStapelImage) toStapelImageBaseGitLayerDirective(name string) (imageBase *StapelImageBase, err error) {
	if len(c.RawGit) != 0 {
		layerName := "git"
		if name != "" {
			layerName = strings.Join([]string{name, layerName}, "-")
		}

		if imageBase, err = c.toBaseStapelImageBaseDirective(layerName); err != nil {
			return nil, err
		}

		imageBase.Git = &GitManager{}
		for _, git := range c.RawGit {
			if git.gitType() == "local" {
				if gitLocal, err := git.toGitLocalDirective(); err != nil {
					return nil, err
				} else {
					imageBase.Git.Local = append(imageBase.Git.Local, gitLocal)
				}
			} else {
				if gitRemote, err := git.toGitRemoteDirective(); err != nil {
					return nil, err
				} else {
					imageBase.Git.Remote = append(imageBase.Git.Remote, gitRemote)
				}
			}
		}
	}

	return
}

func (c *rawStapelImage) toStapelImageBaseImportsLayerDirectiveByBeforeAndAfter(name string, before string, after string) (imageBase *StapelImageBase, err error) {
	if importArtifacts, err := c.layerStapelImportArtifactsByLayer(before, after); err != nil {
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

			if imageBase, err = c.toBaseStapelImageBaseDirective(layerName); err != nil {
				return nil, err
			}
			imageBase.Import = importArtifacts
		} else {
			return nil, nil
		}
	}

	return
}

func (c *rawStapelImage) toStapelImageArtifactTopLayerDirective() (mainImageArtifactLayer *StapelImageArtifact, err error) {
	mainImageArtifactLayer = &StapelImageArtifact{}
	if mainImageArtifactLayer.StapelImageBase, err = c.toBaseStapelImageBaseDirective(c.Artifact); err != nil {
		return nil, err
	}

	return mainImageArtifactLayer, nil
}

func (c *rawStapelImage) layerStapelImportArtifactsByLayer(before string, after string) (artifactImports []*Import, err error) {
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

func (c *rawStapelImage) toShellDirectiveByCommandAndStage(command string, stage string) (shell *Shell) {
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

func (c *rawStapelImage) toStapelImageArtifactAnsibleLayers(tasks []*AnsibleTask, stage string) (imageLayers []*StapelImageArtifact, err error) {
	for ind, task := range tasks {
		layerName := fmt.Sprintf("%s-%s-%d", c.Artifact, strings.ToLower(stage), ind)
		if imageLayer, err := c.toStapelImageArtifactLayerDirective(layerName); err != nil {
			return nil, err
		} else {
			imageLayer.Ansible = c.toAnsibleWithTaskByStage(task, stage)
			imageLayers = append(imageLayers, imageLayer)
		}
	}

	return imageLayers, nil
}

func (c *rawStapelImage) toAnsibleWithTaskByStage(task *AnsibleTask, stage string) (ansible *Ansible) {
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

func (c *rawStapelImage) validateStapelImageArtifactDirective(imageArtifact *StapelImageArtifact) (err error) {
	if c.RawDocker != nil {
		return newDetailedConfigError("`docker` section is not supported for artifact!", nil, c.doc)
	}

	if err := imageArtifact.validate(); err != nil {
		return err
	}

	return nil
}

func (c *rawStapelImage) toStapelImageBaseDirective(name string) (imageBase *StapelImageBase, err error) {
	if imageBase, err = c.toBaseStapelImageBaseDirective(name); err != nil {
		return nil, err
	}

	imageBase.From = c.From
	imageBase.FromImageName = c.FromImage
	imageBase.FromArtifactName = c.FromArtifact
	imageBase.FromLatest = c.FromLatest
	imageBase.FromCacheVersion = c.FromCacheVersion

	for _, git := range c.RawGit {
		if git.gitType() == "local" {
			if gitLocal, err := git.toGitLocalDirective(); err != nil {
				return nil, err
			} else {
				imageBase.Git.Local = append(imageBase.Git.Local, gitLocal)
			}
		} else {
			if gitRemote, err := git.toGitRemoteDirective(); err != nil {
				return nil, err
			} else {
				imageBase.Git.Remote = append(imageBase.Git.Remote, gitRemote)
			}
		}
	}

	if c.RawShell != nil {
		if shell, err := c.RawShell.toDirective(); err != nil {
			return nil, err
		} else {
			imageBase.Shell = shell
		}
	}

	if c.RawAnsible != nil {
		if ansible, err := c.RawAnsible.toDirective(); err != nil {
			return nil, err
		} else {
			imageBase.Ansible = ansible
		}
	}

	for _, importArtifact := range c.RawImport {
		if importArtifactDirective, err := importArtifact.toDirective(); err != nil {
			return nil, err
		} else {
			imageBase.Import = append(imageBase.Import, importArtifactDirective)
		}
	}

	if err := c.validateStapelImageBaseDirective(imageBase); err != nil {
		return nil, err
	}

	return imageBase, nil
}

func (c *rawStapelImage) validateStapelImageBaseDirective(imageBase *StapelImageBase) (err error) {
	if err := imageBase.validate(c.DisableDeterminism); err != nil {
		return err
	}

	return nil
}

func (c *rawStapelImage) toBaseStapelImageBaseDirective(name string) (imageBase *StapelImageBase, err error) {
	imageBase = &StapelImageBase{}
	imageBase.Name = name

	for _, mount := range c.RawMount {
		if imageMount, err := mount.toDirective(); err != nil {
			return nil, err
		} else {
			imageBase.Mount = append(imageBase.Mount, imageMount)
		}
	}

	imageBase.Git = &GitManager{}

	imageBase.raw = c

	return imageBase, nil
}
