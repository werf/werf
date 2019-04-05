package config

import (
	"fmt"
	"strings"
)

type rawImage struct {
	Images            []string     `yaml:"-"`
	Artifact          string       `yaml:"artifact,omitempty"`
	From              string       `yaml:"from,omitempty"`
	FromLatest        bool         `yaml:"fromLatest,omitempty"`
	FromCacheVersion  string       `yaml:"fromCacheVersion,omitempty"`
	FromImage         string       `yaml:"fromImage,omitempty"`
	FromImageArtifact string       `yaml:"fromImageArtifact,omitempty"`
	RawGit            []*rawGit    `yaml:"git,omitempty"`
	RawShell          *rawShell    `yaml:"shell,omitempty"`
	RawAnsible        *rawAnsible  `yaml:"ansible,omitempty"`
	RawMount          []*rawMount  `yaml:"mount,omitempty"`
	RawDocker         *rawDocker   `yaml:"docker,omitempty"`
	RawImport         []*rawImport `yaml:"import,omitempty"`
	AsLayers          bool         `yaml:"asLayers,omitempty"`

	doc *doc `yaml:"-"` // parent

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *rawImage) setAndValidateImage() error {
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

func (c *rawImage) UnmarshalYAML(unmarshal func(interface{}) error) error {
	parentStack.Push(c)
	type plain rawImage
	err := unmarshal((*plain)(c))
	parentStack.Pop()
	if err != nil {
		return err
	}

	if err := c.setAndValidateImage(); err != nil {
		return err
	}

	if err := checkOverflow(c.UnsupportedAttributes, nil, c.doc); err != nil {
		return err
	}

	if err := c.validateImageType(); err != nil {
		return err
	}

	return nil
}

func (c *rawImage) validateImageType() error {
	isImage := len(c.Images) != 0
	isArtifact := c.Artifact != ""

	if isImage && isArtifact {
		return newDetailedConfigError("unknown doc type: one and only one of `image: NAME` or `artifact: NAME` non-empty name required!", nil, c.doc)
	} else if !(isImage || isArtifact) {
		return newDetailedConfigError("unknown doc type: one of `image: NAME` or `artifact: NAME` non-empty name required!", nil, c.doc)
	}

	return nil
}

func (c *rawImage) imageType() string {
	if len(c.Images) != 0 {
		return "images"
	} else if c.Artifact != "" {
		return "artifact"
	}

	return ""
}

func (c *rawImage) toImageDirectives() (images []*Image, err error) {
	for _, imageName := range c.Images {
		if image, err := c.toImageDirective(imageName); err != nil {
			return nil, err
		} else {
			images = append(images, image)
		}
	}

	return images, nil
}

func (c *rawImage) toImageArtifactDirective() (imageArtifact *ImageArtifact, err error) {
	imageArtifact = &ImageArtifact{}
	if c.AsLayers {
		if imageArtifact, err = c.toImageArtifactAsLayersDirective(); err != nil {
			return nil, err
		}
	} else {
		if imageArtifact.ImageBase, err = c.toImageBaseDirective(c.Artifact); err != nil {
			return nil, err
		}
	}

	if err := c.validateArtifactImageDirective(imageArtifact); err != nil {
		return nil, err
	}

	return imageArtifact, nil
}

func (c *rawImage) toImageDirective(name string) (image *Image, err error) {
	image = &Image{}

	if c.AsLayers {
		if image, err = c.toImageAsLayersDirective(name); err != nil {
			return nil, err
		}
	} else {
		if imageBase, err := c.toImageBaseDirective(name); err != nil {
			return nil, err
		} else {
			image.ImageBase = imageBase
		}

		if c.RawDocker != nil {
			if docker, err := c.RawDocker.toDirective(); err != nil {
				return nil, err
			} else {
				image.Docker = docker
			}
		}
	}

	if err := c.validateImageDirective(image); err != nil {
		return nil, err
	}

	return
}

func (c *rawImage) toImageAsLayersDirective(name string) (image *Image, err error) {
	imageBaseLayers, err := c.toImageBaseLayersDirectives(name)
	if err != nil {
		return nil, err
	}

	var layers []*Image
	for _, imageBaseLayer := range imageBaseLayers {
		layer := &Image{}
		layer.ImageBase = imageBaseLayer
		layers = append(layers, layer)
	}

	if image, err = c.toImageTopLayerDirective(name); err != nil {
		return nil, err
	} else {
		layers = append(layers, image)
	}

	var prevImageLayer *Image
	for _, imageLayer := range layers {
		if prevImageLayer == nil {
			imageLayer.From = c.From
			imageLayer.FromLatest = c.FromLatest
			imageLayer.FromCacheVersion = c.FromCacheVersion
		} else {
			imageLayer.FromImage = prevImageLayer
		}
		prevImageLayer = imageLayer
	}

	if err = c.validateImageBaseDirective(prevImageLayer.ImageBase); err != nil {
		return nil, err
	} else {
		return prevImageLayer, nil
	}
}

func (c *rawImage) toImageBaseLayersDirectives(name string) (layers []*ImageBase, err error) {
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
		if beforeInstallShellLayers, err := c.toImageBaseShellLayersDirectivesByStage(name, shell.BeforeInstall, "beforeInstall"); err != nil {
			return nil, err
		} else {
			layers = append(layers, beforeInstallShellLayers...)
		}
	} else if ansible != nil {
		if beforeInstallAnsibleLayers, err := c.toImageBaseAnsibleLayersDirectivesByStage(name, ansible.BeforeInstall, "beforeInstall"); err != nil {
			return nil, err
		} else {
			layers = append(layers, beforeInstallAnsibleLayers...)
		}
	}

	if gitLayer, err := c.toImageBaseGitLayerDirective(name); err != nil {
		return nil, err
	} else if gitLayer != nil {
		layers = append(layers, gitLayer)
	}

	if importsLayers, err := c.toImageBaseImportsLayerDirectiveByBeforeAndAfter(name, "install", ""); err != nil {
		return nil, err
	} else if importsLayers != nil {
		layers = append(layers, importsLayers)
	}

	if shell != nil {
		if installShellLayers, err := c.toImageBaseShellLayersDirectivesByStage(name, shell.Install, "install"); err != nil {
			return nil, err
		} else {
			layers = append(layers, installShellLayers...)
		}
	} else if ansible != nil {
		if installAnsibleLayers, err := c.toImageBaseAnsibleLayersDirectivesByStage(name, ansible.Install, "install"); err != nil {
			return nil, err
		} else {
			layers = append(layers, installAnsibleLayers...)
		}
	}

	if importsLayer, err := c.toImageBaseImportsLayerDirectiveByBeforeAndAfter(name, "", "install"); err != nil {
		return nil, err
	} else if importsLayer != nil {
		layers = append(layers, importsLayer)
	}

	if shell != nil {
		if beforeSetupShellLayers, err := c.toImageBaseShellLayersDirectivesByStage(name, shell.BeforeSetup, "beforeSetup"); err != nil {
			return nil, err
		} else {
			layers = append(layers, beforeSetupShellLayers...)
		}
	} else if ansible != nil {
		if beforeSetupAnsibleLayers, err := c.toImageBaseAnsibleLayersDirectivesByStage(name, ansible.BeforeSetup, "beforeSetup"); err != nil {
			return nil, err
		} else {
			layers = append(layers, beforeSetupAnsibleLayers...)
		}
	}

	if importsLayer, err := c.toImageBaseImportsLayerDirectiveByBeforeAndAfter(name, "setup", ""); err != nil {
		return nil, err
	} else if importsLayer != nil {
		layers = append(layers, importsLayer)
	}

	if shell != nil {
		if setupShellLayers, err := c.toImageBaseShellLayersDirectivesByStage(name, shell.Setup, "setup"); err != nil {
			return nil, err
		} else {
			layers = append(layers, setupShellLayers...)
		}
	} else if ansible != nil {
		if setupAnsibleLayers, err := c.toImageBaseAnsibleLayersDirectivesByStage(name, ansible.Setup, "setup"); err != nil {
			return nil, err
		} else {
			layers = append(layers, setupAnsibleLayers...)
		}
	}

	return layers, nil
}

func (c *rawImage) toImageBaseShellLayersDirectivesByStage(name string, commands []string, stage string) (imageLayers []*ImageBase, err error) {
	for ind, command := range commands {
		layerName := fmt.Sprintf("%s-%d", strings.ToLower(stage), ind)
		if name != "" {
			layerName = strings.Join([]string{name, layerName}, "-")
		}

		if imageBaseLayer, err := c.toBaseImageBaseDirective(layerName); err != nil {
			return nil, err
		} else {
			imageBaseLayer.Shell = c.toShellDirectiveByCommandAndStage(command, stage)
			imageLayers = append(imageLayers, imageBaseLayer)
		}
	}

	return imageLayers, nil
}

func (c *rawImage) toImageBaseAnsibleLayersDirectivesByStage(name string, tasks []*AnsibleTask, stage string) (layers []*ImageBase, err error) {
	for ind, task := range tasks {
		layerName := fmt.Sprintf("%s-%d", strings.ToLower(stage), ind)
		if name != "" {
			layerName = strings.Join([]string{name, layerName}, "-")
		}

		if layer, err := c.toBaseImageBaseDirective(layerName); err != nil {
			return nil, err
		} else {
			layer.Ansible = c.toAnsibleWithTaskByStage(task, stage)
			layers = append(layers, layer)
		}
	}

	return layers, nil
}

func (c *rawImage) toImageLayerDirective(layerName string) (image *Image, err error) {
	image = &Image{}
	if image.ImageBase, err = c.toBaseImageBaseDirective(layerName); err != nil {
		return nil, err
	}
	return
}

func (c *rawImage) toImageTopLayerDirective(name string) (mainImageLayer *Image, err error) {
	mainImageLayer = &Image{}
	if mainImageLayer.ImageBase, err = c.toBaseImageBaseDirective(name); err != nil {
		return nil, err
	}

	if mainImageLayer.Import, err = c.layerImportArtifactsByLayer("", "setup"); err != nil {
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

func (c *rawImage) validateImageDirective(image *Image) (err error) {
	if err := image.validate(); err != nil {
		return err
	}

	return nil
}

func (c *rawImage) toImageArtifactAsLayersDirective() (imageArtifactLayer *ImageArtifact, err error) {
	imageBaseLayers, err := c.toImageBaseLayersDirectives(c.Artifact)
	if err != nil {
		return nil, err
	}

	var layers []*ImageArtifact
	for _, imageBaseLayer := range imageBaseLayers {
		layer := &ImageArtifact{}
		layer.ImageBase = imageBaseLayer
		layers = append(layers, layer)
	}

	if imageArtifactLayer, err = c.toImageArtifactTopLayerDirective(); err != nil {
		return nil, err
	} else {
		layers = append(layers, imageArtifactLayer)
	}

	var prevImageLayer *ImageArtifact
	for _, layer := range layers {
		if prevImageLayer == nil {
			layer.From = c.From
			layer.FromLatest = c.FromLatest
			layer.FromCacheVersion = c.FromCacheVersion
		} else {
			layer.FromImageArtifact = prevImageLayer
		}

		prevImageLayer = layer
	}

	if err = c.validateImageBaseDirective(prevImageLayer.ImageBase); err != nil {
		return nil, err
	} else {
		return prevImageLayer, nil
	}
}

func (c *rawImage) toImageArtifactLayerDirective(layerName string) (imageArtifact *ImageArtifact, err error) {
	imageArtifact = &ImageArtifact{}
	if imageArtifact.ImageBase, err = c.toBaseImageBaseDirective(layerName); err != nil {
		return nil, err
	}
	return
}

func (c *rawImage) toImageArtifactLayerWithGitDirective() (imageArtifact *ImageArtifact, err error) {
	if imageBase, err := c.toImageBaseGitLayerDirective(c.Artifact); err == nil && imageBase != nil {
		imageArtifact = &ImageArtifact{}
		imageArtifact.ImageBase = imageBase
	}
	return
}

func (c *rawImage) toImageArtifactLayerWithArtifactsDirective(before string, after string) (imageArtifact *ImageArtifact, err error) {
	if imageBase, err := c.toImageBaseImportsLayerDirectiveByBeforeAndAfter(c.Artifact, before, after); err == nil && imageBase != nil {
		imageArtifact = &ImageArtifact{}
		imageArtifact.ImageBase = imageBase
	}
	return
}

func (c *rawImage) toImageBaseGitLayerDirective(name string) (imageBase *ImageBase, err error) {
	if len(c.RawGit) != 0 {
		layerName := "git"
		if name != "" {
			layerName = strings.Join([]string{name, layerName}, "-")
		}

		if imageBase, err = c.toBaseImageBaseDirective(layerName); err != nil {
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

func (c *rawImage) toImageBaseImportsLayerDirectiveByBeforeAndAfter(name string, before string, after string) (imageBase *ImageBase, err error) {
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

			if imageBase, err = c.toBaseImageBaseDirective(layerName); err != nil {
				return nil, err
			}
			imageBase.Import = importArtifacts
		} else {
			return nil, nil
		}
	}

	return
}

func (c *rawImage) toImageArtifactTopLayerDirective() (mainImageArtifactLayer *ImageArtifact, err error) {
	mainImageArtifactLayer = &ImageArtifact{}
	if mainImageArtifactLayer.ImageBase, err = c.toBaseImageBaseDirective(c.Artifact); err != nil {
		return nil, err
	}

	return mainImageArtifactLayer, nil
}

func (c *rawImage) layerImportArtifactsByLayer(before string, after string) (artifactImports []*Import, err error) {
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

func (c *rawImage) toShellDirectiveByCommandAndStage(command string, stage string) (shell *Shell) {
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

func (c *rawImage) toImageArtifactAnsibleLayers(tasks []*AnsibleTask, stage string) (imageLayers []*ImageArtifact, err error) {
	for ind, task := range tasks {
		layerName := fmt.Sprintf("%s-%s-%d", c.Artifact, strings.ToLower(stage), ind)
		if imageLayer, err := c.toImageArtifactLayerDirective(layerName); err != nil {
			return nil, err
		} else {
			imageLayer.Ansible = c.toAnsibleWithTaskByStage(task, stage)
			imageLayers = append(imageLayers, imageLayer)
		}
	}

	return imageLayers, nil
}

func (c *rawImage) toAnsibleWithTaskByStage(task *AnsibleTask, stage string) (ansible *Ansible) {
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

func (c *rawImage) validateArtifactImageDirective(imageArtifact *ImageArtifact) (err error) {
	if c.RawDocker != nil {
		return newDetailedConfigError("`docker` section is not supported for artifact!", nil, c.doc)
	}

	if err := imageArtifact.validate(); err != nil {
		return err
	}

	return nil
}

func (c *rawImage) toImageBaseDirective(name string) (imageBase *ImageBase, err error) {
	if imageBase, err = c.toBaseImageBaseDirective(name); err != nil {
		return nil, err
	}

	imageBase.From = c.From
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

	if err := c.validateImageBaseDirective(imageBase); err != nil {
		return nil, err
	}

	return imageBase, nil
}

func (c *rawImage) validateImageBaseDirective(imageBase *ImageBase) (err error) {
	if err := imageBase.validate(); err != nil {
		return err
	}

	return nil
}

func (c *rawImage) toBaseImageBaseDirective(name string) (imageBase *ImageBase, err error) {
	imageBase = &ImageBase{}
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
