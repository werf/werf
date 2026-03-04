package config

import (
	"fmt"

	"github.com/werf/werf/v2/pkg/giterminism_manager"
	"github.com/werf/werf/v2/pkg/util/option"
)

type rawStapelImage struct {
	Images               []string         `yaml:"-"`
	Final                *bool            `yaml:"final,omitempty"`
	CacheVersion         string           `yaml:"cacheVersion,omitempty"`
	From                 string           `yaml:"from,omitempty"`
	FromLatest           bool             `yaml:"fromLatest,omitempty"`
	FromCacheVersion     string           `yaml:"fromCacheVersion,omitempty"`
	FromImage            string           `yaml:"fromImage,omitempty"`
	DisableGitAfterPatch bool             `yaml:"disableGitAfterPatch,omitempty"`
	RawGit               []*rawGit        `yaml:"git,omitempty"`
	RawShell             *rawShell        `yaml:"shell,omitempty"`
	RawMount             []*rawMount      `yaml:"mount,omitempty"`
	RawImport            []*rawImport     `yaml:"import,omitempty"`
	RawDependencies      []*rawDependency `yaml:"dependencies,omitempty"`
	Platform             []string         `yaml:"platform,omitempty"`
	Network              string           `yaml:"network,omitempty"`
	RawSecrets           []*rawSecret     `yaml:"secrets,omitempty"`
	RawImageSpec         *rawImageSpec    `yaml:"imageSpec,omitempty"`

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

	if !isImage {
		return newDetailedConfigError("unknown doc type: `image: NAME` is required!", nil, c.doc)
	}

	return nil
}

func (c *rawStapelImage) stapelImageType() string {
	if len(c.Images) != 0 {
		return "images"
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

func (c *rawStapelImage) toStapelImageDirective(giterminismManager giterminism_manager.Interface, name string) (*StapelImage, error) {
	image := &StapelImage{}

	if imageBase, err := c.toStapelImageBaseDirective(giterminismManager, name); err != nil {
		return nil, err
	} else {
		image.StapelImageBase = imageBase
	}

	image.StapelImageBase.final = option.PtrValueOrDefault(c.Final, true)

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

func (c *rawStapelImage) toStapelImageBaseDirective(giterminismManager giterminism_manager.Interface, name string) (imageBase *StapelImageBase, err error) {
	if imageBase, err = c.toBaseStapelImageBaseDirective(giterminismManager, name); err != nil {
		return nil, err
	}

	imageBase.From = c.From
	if c.FromImage != "" {
		imageBase.From = c.FromImage
	}

	imageBase.FromLatest = c.FromLatest
	imageBase.FromCacheVersion = c.FromCacheVersion

	// TODO(major): This is a dirty temporary backward compatibility fix. Remove it.
	if imageBase.Name == imageBase.From {
		imageBase.From += ":latest"
	}

	imageBase.cacheVersion = c.CacheVersion
	imageBase.platform = append([]string{}, c.Platform...)
	imageBase.Network = c.Network

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

	for _, importArtifact := range c.RawImport {
		if importArtifactDirective, err := importArtifact.toDirective(); err != nil {
			return nil, err
		} else {
			imageBase.Import = append(imageBase.Import, importArtifactDirective)
		}
	}

	if err := imageBase.exportsAutoExcluding(); err != nil {
		return nil, err
	}

	for _, rawDep := range c.RawDependencies {
		dependencyDirective, err := rawDep.toDirective()
		if err != nil {
			return nil, err
		}

		imageBase.Dependencies = append(imageBase.Dependencies, dependencyDirective)
	}

	secrets, err := GetValidatedSecrets(c.RawSecrets, giterminismManager, c.doc)
	if err != nil {
		return nil, err
	}

	imageBase.Secrets = secrets

	if c.RawImageSpec != nil {
		imageBase.ImageSpec = c.RawImageSpec.toDirective()
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
	imageBase.Git.isGitAfterPatchDisabled = c.DisableGitAfterPatch

	imageBase.raw = c

	return imageBase, nil
}

func (r *rawStapelImage) getDoc() *doc {
	return r.doc
}
