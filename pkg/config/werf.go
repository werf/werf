package config

import (
	"errors"
	"fmt"
	"strings"
)

type WerfConfig struct {
	Meta                 *Meta
	StapelImages         []*StapelImage
	ImagesFromDockerfile []*ImageFromDockerfile
	Artifacts            []*StapelImageArtifact
}

func (c *WerfConfig) HasImageOrArtifact(imageName string) bool {
	if c.HasImage(imageName) || c.GetArtifact(imageName) != nil {
		return true
	}

	return false
}

func (c *WerfConfig) HasImage(imageName string) bool {
	if c.GetImage(imageName) != nil {
		return true
	}

	return false
}

func (c *WerfConfig) GetAllImages() []ImageInterface {
	var images []ImageInterface

	for _, image := range c.StapelImages {
		images = append(images, image)
	}

	for _, image := range c.ImagesFromDockerfile {
		images = append(images, image)
	}

	return images
}

func (c *WerfConfig) GetImage(imageName string) ImageInterface {
	if i := c.GetStapelImage(imageName); i != nil {
		return i
	}

	if i := c.GetDockerfileImage(imageName); i != nil {
		return i
	}

	return nil
}

func (c *WerfConfig) GetStapelImage(imageName string) *StapelImage {
	for _, image := range c.StapelImages {
		if image.Name == imageName {
			return image
		}
	}

	return nil
}

func (c *WerfConfig) GetDockerfileImage(imageName string) *ImageFromDockerfile {
	for _, image := range c.ImagesFromDockerfile {
		if image.Name == imageName {
			return image
		}
	}

	return nil
}

func (c *WerfConfig) GetArtifact(imageName string) *StapelImageArtifact {
	for _, artifact := range c.Artifacts {
		if artifact.Name == imageName {
			return artifact
		}
	}

	return nil
}

func (c *WerfConfig) exportsAutoExcluding() error {
	for _, image := range c.StapelImages {
		if err := image.exportsAutoExcluding(); err != nil {
			return err
		}
	}

	for _, artifact := range c.Artifacts {
		if err := artifact.exportsAutoExcluding(); err != nil {
			return err
		}
	}

	return nil
}

func (c *WerfConfig) validateImagesNames() error {
	imageByName := map[string]ImageInterface{}
	for _, image := range c.StapelImages {
		name := image.Name

		if name == "" && (len(c.StapelImages) > 1 || len(c.ImagesFromDockerfile) > 1) {
			return newConfigError(fmt.Sprintf("conflict between images names: a nameless image cannot be specified in the config with multiple images!\n\n%s\n", dumpConfigDoc(image.raw.doc)))
		}

		if d, ok := imageByName[name]; ok {
			return newConfigError(fmt.Sprintf("conflict between images names!\n\n%s%s\n", dumpConfigDoc(d.(*StapelImage).raw.doc), dumpConfigDoc(image.raw.doc)))
		} else {
			imageByName[name] = image
		}
	}

	for _, image := range c.ImagesFromDockerfile {
		name := image.Name

		if name == "" && (len(c.StapelImages) > 1 || len(c.ImagesFromDockerfile) > 1) {
			return newConfigError(fmt.Sprintf("conflict between images names: a nameless image cannot be specified in the config with multiple images!\n\n%s\n", dumpConfigDoc(image.raw.doc)))
		}

		if d, ok := imageByName[name]; ok {
			var doc string
			switch i := d.(type) {
			case *StapelImage:
				doc = dumpConfigDoc(i.raw.doc)
			case *ImageFromDockerfile:
				doc = dumpConfigDoc(i.raw.doc)
			}

			return newConfigError(fmt.Sprintf("conflict between images names!\n\n%s%s\n", doc, dumpConfigDoc(image.raw.doc)))
		} else {
			imageByName[name] = image
		}
	}

	imageArtifactByName := map[string]*StapelImageArtifact{}
	for _, artifact := range c.Artifacts {
		name := artifact.Name

		if a, ok := imageArtifactByName[name]; ok {
			return newConfigError(fmt.Sprintf("conflict between artifacts names!\n\n%s%s\n", dumpConfigDoc(a.raw.doc), dumpConfigDoc(artifact.raw.doc)))
		} else {
			imageArtifactByName[name] = artifact
		}

		if iInterface, exist := imageByName[name]; exist {
			var doc string
			switch i := iInterface.(type) {
			case *StapelImage:
				doc = dumpConfigDoc(i.raw.doc)
			case *ImageFromDockerfile:
				doc = dumpConfigDoc(i.raw.doc)
			}
			return newConfigError(fmt.Sprintf("conflict between image and artifact names!\n\n%s%s\n", doc, dumpConfigDoc(artifact.raw.doc)))
		} else {
			imageArtifactByName[name] = artifact
		}
	}

	return nil
}

func (c *WerfConfig) associateImportsArtifacts() error {
	var relatedImageImages []ImageInterface
	var artifactImports []*Import

	for _, image := range c.StapelImages {
		relatedImageImages = append(relatedImageImages, c.relatedImageImages(image)...)
	}

	for _, artifact := range c.Artifacts {
		relatedImageImages = append(relatedImageImages, c.relatedImageImages(artifact)...)
	}

	for _, relatedImageInterface := range relatedImageImages {
		switch relatedImageInterface.(type) {
		case *StapelImage:
			artifactImports = append(artifactImports, relatedImageInterface.(*StapelImage).Import...)
		case *StapelImageArtifact:
			artifactImports = append(artifactImports, relatedImageInterface.(*StapelImageArtifact).Import...)
		}
	}

	for _, artifactImport := range artifactImports {
		if err := c.validateImportImage(artifactImport); err != nil {
			return err
		}
	}

	return nil
}

func (c *WerfConfig) validateImportImage(i *Import) error {
	if i.ImageName != "" {
		if interf := c.GetImage(i.ImageName); interf == nil {
			var imageName string

			switch image := interf.(type) {
			case *StapelImage:
				imageName = image.Name
			case *ImageFromDockerfile:
				imageName = image.Name
			}

			return newDetailedConfigError(fmt.Sprintf("no such image `%s`!", imageName), i.raw, i.raw.rawStapelImage.doc)
		}
	} else if i.ArtifactName != "" {
		if imageArtifact := c.GetArtifact(i.ArtifactName); imageArtifact == nil {
			return newDetailedConfigError(fmt.Sprintf("no such artifact `%s`!", i.ArtifactName), i.raw, i.raw.rawStapelImage.doc)
		}
	}

	return nil
}

func (c *WerfConfig) validateImagesFrom() error {
	for _, image := range c.StapelImages {
		if err := c.validateImageFrom(image.StapelImageBase); err != nil {
			return err
		}
	}

	for _, image := range c.Artifacts {
		if err := c.validateImageFrom(image.StapelImageBase); err != nil {
			return err
		}
	}

	return nil
}

func (c *WerfConfig) validateImageFrom(i *StapelImageBase) error {
	if i.raw.FromImage != "" {
		fromImageName := i.raw.FromImage

		if fromImageName == i.Name {
			return newDetailedConfigError(fmt.Sprintf("cannot use own image name as `fromImage` directive value!"), nil, i.raw.doc)
		}

		if interf := c.GetImage(fromImageName); interf == nil {
			return newDetailedConfigError(fmt.Sprintf("no such image `%s`!", fromImageName), i.raw, i.raw.doc)
		}
	} else if i.raw.FromImageArtifact != "" {
		fromImageArtifactName := i.raw.FromImageArtifact

		if fromImageArtifactName == i.Name {
			return newDetailedConfigError(fmt.Sprintf("cannot use own image name as `fromImageArtifact` directive value!"), nil, i.raw.doc)
		}

		if imageArtifact := c.GetArtifact(fromImageArtifactName); imageArtifact == nil {
			return newDetailedConfigError(fmt.Sprintf("no such image artifact `%s`!", fromImageArtifactName), i.raw, i.raw.doc)
		}
	}

	return nil
}

func (c *WerfConfig) validateInfiniteLoopBetweenRelatedImages() error {
	var imageAndArtifactNames []string

	for _, image := range c.StapelImages {
		imageAndArtifactNames = append(imageAndArtifactNames, image.Name)
	}

	for _, artifact := range c.Artifacts {
		imageAndArtifactNames = append(imageAndArtifactNames, artifact.Name)
	}

	for _, imageOrArtifactName := range imageAndArtifactNames {
		if err, errImagesStack := c.validateImageInfiniteLoop(imageOrArtifactName, []string{}); err != nil {
			return fmt.Errorf("%s: %s", err, strings.Join(errImagesStack, " -> "))
		}
	}

	return nil
}

func (c *WerfConfig) ImageTree(interf ImageInterface) (tree []ImageInterface) {
	switch i := interf.(type) {
	case StapelImageInterface:
		if i.ImageBaseConfig().FromImageName != "" {
			tree = append(tree, c.ImageTree(c.GetImage(i.ImageBaseConfig().FromImageName))...)
		}

		if i.ImageBaseConfig().FromImageArtifactName != "" {
			tree = append(tree, c.ImageTree(c.GetArtifact(i.ImageBaseConfig().FromImageArtifactName))...)
		}

		for _, imp := range i.imports() {
			if imp.ImageName != "" {
				tree = append(tree, c.ImageTree(c.GetImage(imp.ImageName))...)
			} else if imp.ArtifactName != "" {
				tree = append(tree, c.ImageTree(c.GetArtifact(imp.ArtifactName))...)
			}
		}
	case *ImageFromDockerfile:
	}

	return append(tree, interf)
}

func (c *WerfConfig) relatedImageImages(interf ImageInterface) (images []ImageInterface) {
	images = append(images, interf)
	switch i := interf.(type) {
	case StapelImageInterface:
		if i.ImageBaseConfig().FromImageName != "" {
			images = append(images, c.relatedImageImages(c.GetImage(i.ImageBaseConfig().FromImageName))...)
		}

		if i.ImageBaseConfig().FromImageArtifactName != "" {
			images = append(images, c.relatedImageImages(c.GetArtifact(i.ImageBaseConfig().FromImageArtifactName))...)
		}
	case *ImageFromDockerfile:
	}

	return
}

func (c *WerfConfig) validateImageInfiniteLoop(imageOrArtifactName string, imageNameStack []string) (error, []string) {
	for _, stackImageName := range imageNameStack {
		if stackImageName == imageOrArtifactName {
			return errors.New("infinite loop detected"), []string{imageOrArtifactName}
		}
	}
	imageNameStack = append(imageNameStack, imageOrArtifactName)

	var image StapelImageInterface
	interf := c.GetImage(imageOrArtifactName)
	if interf != nil {
		switch i := interf.(type) {
		case *ImageFromDockerfile:
			return nil, imageNameStack
		case *StapelImage:
			image = i
		}
	} else {
		image = c.GetArtifact(imageOrArtifactName)
		if image == nil {
			panic("runtime error")
		}
	}

	imageBaseConfig := image.ImageBaseConfig()
	if imageBaseConfig == nil {
		panic("runtime error")
	}

	if imageBaseConfig.FromImageName != "" {
		if err, errImagesStack := c.validateImageInfiniteLoop(imageBaseConfig.FromImageName, imageNameStack); err != nil {
			return err, append([]string{imageOrArtifactName}, errImagesStack...)
		}
	}

	if imageBaseConfig.FromImageArtifactName != "" {
		if err, errImagesStack := c.validateImageInfiniteLoop(imageBaseConfig.FromImageArtifactName, imageNameStack); err != nil {
			return err, append([]string{imageOrArtifactName}, errImagesStack...)
		}
	}

	for _, imp := range image.imports() {
		if imp.ImageName != "" {
			var importImageName string

			switch i := c.GetImage(imp.ImageName).(type) {
			case *ImageFromDockerfile:
				importImageName = i.Name
			case *StapelImage:
				importImageName = i.Name
			}

			if err, errImagesStack := c.validateImageInfiniteLoop(importImageName, imageNameStack); err != nil {
				return err, append([]string{imageOrArtifactName}, errImagesStack...)
			}
		} else if imp.ArtifactName != "" {
			if err, errImagesStack := c.validateImageInfiniteLoop(imp.ArtifactName, imageNameStack); err != nil {
				return err, append([]string{imageOrArtifactName}, errImagesStack...)
			}
		}
	}

	return nil, nil
}
