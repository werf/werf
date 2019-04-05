package config

import (
	"fmt"
)

type ImageBase struct {
	Name              string
	From              string
	FromLatest        bool
	FromImage         *Image
	FromImageArtifact *ImageArtifact
	FromCacheVersion  string
	Git               *GitManager
	Shell             *Shell
	Ansible           *Ansible
	Mount             []*Mount
	Import            []*Import

	raw *rawImage
}

func (c *ImageBase) fromImage() *Image {
	return c.FromImage
}

func (c *ImageBase) fromImageArtifact() *ImageArtifact {
	return c.FromImageArtifact
}

func (c *ImageBase) imports() []*Import {
	return c.Import
}

func (c *ImageBase) ImageBaseConfig() *ImageBase {
	return c
}

func (c *ImageBase) IsArtifact() bool {
	return false
}

func (c *ImageBase) exportsAutoExcluding() error {
	for _, exp1 := range c.exports() {
		for _, exp2 := range c.exports() {
			if exp1 == exp2 {
				continue
			}

			if !exp1.AutoExcludeExportAndCheck(exp2) {
				errMsg := fmt.Sprintf("Conflict between imports!\n\n%s\n%s", dumpConfigSection(exp1.GetRaw()), dumpConfigSection(exp2.GetRaw()))
				return newDetailedConfigError(errMsg, nil, c.raw.doc)
			}
		}
	}

	return nil
}

func (c *ImageBase) exports() []autoExcludeExport {
	var exports []autoExcludeExport
	if c.Git != nil {
		for _, git := range c.Git.Local {
			exports = append(exports, git)
		}

		for _, git := range c.Git.Remote {
			exports = append(exports, git)
		}
	}

	for _, imp := range c.Import {
		exports = append(exports, imp)
	}

	return exports
}

func (c *ImageBase) associateFrom(images []*Image, artifacts []*ImageArtifact) error {
	if c.FromImage != nil || c.FromImageArtifact != nil { // asLayers
		return nil
	}

	if c.raw.FromImage != "" {
		fromImageName := c.raw.FromImage

		if fromImageName == c.Name {
			return newDetailedConfigError(fmt.Sprintf("cannot use own image name as `fromImage` directive value!"), nil, c.raw.doc)
		}

		if image := imageByName(images, fromImageName); image != nil {
			c.FromImage = image
		} else {
			return newDetailedConfigError(fmt.Sprintf("no such image `%s`!", fromImageName), c.raw, c.raw.doc)
		}
	} else if c.raw.FromImageArtifact != "" {
		fromImageArtifactName := c.raw.FromImageArtifact

		if fromImageArtifactName == c.Name {
			return newDetailedConfigError(fmt.Sprintf("cannot use own image name as `fromImageArtifact` directive value!"), nil, c.raw.doc)
		}

		if imageArtifact := imageArtifactByName(artifacts, fromImageArtifactName); imageArtifact != nil {
			c.FromImageArtifact = imageArtifact
		} else {
			return newDetailedConfigError(fmt.Sprintf("no such image artifact `%s`!", fromImageArtifactName), c.raw, c.raw.doc)
		}
	}

	return nil
}

func imageByName(images []*Image, name string) *Image {
	for _, image := range images {
		if image.Name == name {
			return image
		}
	}
	return nil
}

func imageArtifactByName(images []*ImageArtifact, name string) *ImageArtifact {
	for _, image := range images {
		if image.Name == name {
			return image
		}
	}
	return nil
}

func (c *ImageBase) validate() error {
	if c.From == "" && c.raw.FromImage == "" && c.raw.FromImageArtifact == "" && c.FromImage == nil && c.FromImageArtifact == nil {
		return newDetailedConfigError("`from: DOCKER_IMAGE`, `fromImage: IMAGE_NAME`, `fromImageArtifact: IMAGE_ARTIFACT_NAME` required!", nil, c.raw.doc)
	}

	mountByTo := map[string]bool{}
	for _, mount := range c.Mount {
		_, exist := mountByTo[mount.To]
		if exist {
			return newDetailedConfigError("conflict between mounts!", nil, c.raw.doc)
		}

		mountByTo[mount.To] = true
	}

	if !oneOrNone([]bool{c.From != "", c.raw.FromImage != "", c.raw.FromImageArtifact != ""}) {
		return newDetailedConfigError("conflict between `from`, `fromImage` and `fromImageArtifact` directives!", nil, c.raw.doc)
	}

	// TODO: валидацию формата `From`
	// TODO: валидация формата `Name`

	return nil
}
