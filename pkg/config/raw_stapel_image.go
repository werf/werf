package config

import (
	"fmt"

	"github.com/werf/werf/pkg/giterminism_manager"
)

type rawStapelImage struct {
	Images           []string         `yaml:"-"`
	Artifact         string           `yaml:"artifact,omitempty"`
	From             string           `yaml:"from,omitempty"`
	FromLatest       bool             `yaml:"fromLatest,omitempty"`
	FromCacheVersion string           `yaml:"fromCacheVersion,omitempty"`
	FromImage        string           `yaml:"fromImage,omitempty"`
	FromArtifact     string           `yaml:"fromArtifact,omitempty"`
	RawGit           []*rawGit        `yaml:"git,omitempty"`
	RawShell         *rawShell        `yaml:"shell,omitempty"`
	RawAnsible       *rawAnsible      `yaml:"ansible,omitempty"`
	RawMount         []*rawMount      `yaml:"mount,omitempty"`
	RawDocker        *rawDocker       `yaml:"docker,omitempty"`
	RawImport        []*rawImport     `yaml:"import,omitempty"`
	RawDependencies  []*rawDependency `yaml:"dependencies,omitempty"`
	Platform         []string         `yaml:"platform,omitempty"`

	doc *doc `yaml:"-"` // parent

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

func (c *rawStapelImage) toStapelImageDirectives(giterminismManager giterminism_manager.Interface) (images []*StapelImage, err error) {
	for _, imageName := range c.Images {
		if image, err := c.toStapelImageDirective(giterminismManager, imageName); err != nil {
			return nil, err
		} else {
			images = append(images, image)
		}
	}

	return images, nil
}

func (c *rawStapelImage) toStapelImageArtifactDirectives(giterminismManager giterminism_manager.Interface) (*StapelImageArtifact, error) {
	imageArtifact := &StapelImageArtifact{}

	var err error
	if imageArtifact.StapelImageBase, err = c.toStapelImageBaseDirective(giterminismManager, c.Artifact, true); err != nil {
		return nil, err
	}

	if err := c.validateStapelImageArtifactDirective(imageArtifact); err != nil {
		return nil, err
	}

	return imageArtifact, nil
}

func (c *rawStapelImage) toStapelImageDirective(giterminismManager giterminism_manager.Interface, name string) (*StapelImage, error) {
	image := &StapelImage{}

	if imageBase, err := c.toStapelImageBaseDirective(giterminismManager, name, false); err != nil {
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

	if err := c.validateStapelImageDirective(image); err != nil {
		return nil, err
	}

	return image, nil
}

func (c *rawStapelImage) validateStapelImageDirective(image *StapelImage) (err error) {
	if err := image.validate(); err != nil {
		return err
	}

	return nil
}

//nolint:unused
func (c *rawStapelImage) toShellDirectiveByCommandAndStage(command, stage string) (shell *Shell) {
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

//nolint:unused
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

func (c *rawStapelImage) toStapelImageBaseDirective(giterminismManager giterminism_manager.Interface, name string, isArtifact bool) (imageBase *StapelImageBase, err error) {
	if imageBase, err = c.toBaseStapelImageBaseDirective(giterminismManager, name); err != nil {
		return nil, err
	}

	imageBase.From = c.From
	imageBase.FromImageName = c.FromImage
	imageBase.FromArtifactName = c.FromArtifact
	imageBase.FromLatest = c.FromLatest
	imageBase.FromCacheVersion = c.FromCacheVersion
	imageBase.Platform = append([]string{}, c.Platform...)

	for _, git := range c.RawGit {
		if git.gitType() == "local" {
			if gitLocal, err := git.toGitLocalDirective(); err != nil {
				return nil, err
			} else {
				imageBase.Git.Local = append(imageBase.Git.Local, gitLocal)
			}
		} else {
			if gitRemote, err := git.toGitRemoteDirective(giterminismManager); err != nil {
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

	if isArtifact && len(c.RawDependencies) > 0 {
		return nil, newDetailedConfigError(fmt.Sprintf("dependencies directive is specified for %q artifact, but dependencies are not supported for artifacts!", name), nil, c.doc)
	}

	for _, rawDep := range c.RawDependencies {
		dependencyDirective, err := rawDep.toDirective()
		if err != nil {
			return nil, err
		}

		imageBase.Dependencies = append(imageBase.Dependencies, dependencyDirective)
	}

	if err := c.validateStapelImageBaseDirective(giterminismManager, imageBase); err != nil {
		return nil, err
	}

	return imageBase, nil
}

func (c *rawStapelImage) validateStapelImageBaseDirective(giterminismManager giterminism_manager.Interface, imageBase *StapelImageBase) (err error) {
	if err := imageBase.validate(giterminismManager); err != nil {
		return err
	}

	return nil
}

func (c *rawStapelImage) toBaseStapelImageBaseDirective(giterminismManager giterminism_manager.Interface, name string) (imageBase *StapelImageBase, err error) {
	imageBase = &StapelImageBase{}
	imageBase.Name = name

	for _, mount := range c.RawMount {
		if imageMount, err := mount.toDirective(giterminismManager); err != nil {
			return nil, err
		} else {
			imageBase.Mount = append(imageBase.Mount, imageMount)
		}
	}

	imageBase.Git = &GitManager{}

	imageBase.raw = c

	return imageBase, nil
}
