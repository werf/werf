package config

import (
	"errors"
	"fmt"
	"strings"
)

type WerfConfig struct {
	Meta        *Meta
	images      []ImageInterface
	imagesCache map[string]ImageInterface
}

func NewWerfConfig(meta *Meta, images []ImageInterface) *WerfConfig {
	return &WerfConfig{
		Meta:        meta,
		images:      images,
		imagesCache: make(map[string]ImageInterface),
	}
}

func (c *WerfConfig) getSpecificImages(imagesToProcess ImagesToProcess) []ImageInterface {
	var resultImages []ImageInterface
	for _, name := range imagesToProcess.ImageNameList {
		image := c.GetImage(name)
		if image == nil {
			panic(fmt.Sprintf("image %q not found but must be found", name))
		}

		resultImages = append(resultImages, image)
	}

	return resultImages
}

func (c *WerfConfig) Images(onlyFinal bool) []ImageInterface {
	var resultImages []ImageInterface
	for _, image := range c.images {
		if onlyFinal && !image.IsFinal() {
			continue
		}
		resultImages = append(resultImages, image)
	}

	return resultImages
}

func (c *WerfConfig) GetImageNameList(onlyFinal bool) []string {
	var list []string
	for _, image := range c.Images(onlyFinal) {
		list = append(list, image.GetName())
	}

	return list
}

func (c *WerfConfig) GetImage(imageName string) ImageInterface {
	if c.imagesCache == nil {
		c.imagesCache = make(map[string]ImageInterface)
	}
	if v, ok := c.imagesCache[imageName]; ok {
		return v
	}
	for _, image := range c.Images(false) {
		if image.GetName() == imageName {
			c.imagesCache[imageName] = image
			return image
		}
	}

	return nil
}

func (c *WerfConfig) validateConflictBetweenImagesNames() error {
	imageByName := map[string]ImageInterface{}
	for _, image := range c.Images(false) {
		name := image.GetName()
		if name == "" && len(c.Images(true)) > 1 {
			return newConfigError(fmt.Sprintf("conflict between images names: a nameless image cannot be specified in the config with multiple images!\n\n%s\n", dumpConfigDoc(image.rawDoc())))
		}

		if d, ok := imageByName[name]; ok {
			return newConfigError(fmt.Sprintf("conflict between images names!\n\n%s%s\n", dumpConfigDoc(d.rawDoc()), dumpConfigDoc(image.rawDoc())))
		} else {
			imageByName[name] = image
		}
	}

	return nil
}

func (c *WerfConfig) validateRelatedImages() error {
	for _, image := range c.Images(false) {
		for _, relatedImageName := range image.dependsOn().relatedImageNameList() {
			if c.GetImage(relatedImageName) == nil {
				return newDetailedConfigError(fmt.Sprintf("image %q not found!", relatedImageName), nil, image.rawDoc())
			}
		}
	}

	return nil
}

func (c *WerfConfig) validateInfiniteLoopBetweenRelatedImages() error {
	for _, image := range c.Images(false) {
		if err, errImageStack := c.validateImageInfiniteLoop(image.GetName(), []string{}); err != nil {
			return fmt.Errorf("%w: %s", err, strings.Join(errImageStack, " -> "))
		}
	}

	return nil
}

func (c *WerfConfig) validateImageInfiniteLoop(imageName string, imageNameStack []string) (error, []string) {
	for _, stackImageName := range imageNameStack {
		if stackImageName == imageName {
			return errors.New("infinite loop detected"), []string{imageName}
		}
	}
	imageNameStack = append(imageNameStack, imageName)

	image := c.GetImage(imageName)
	if image == nil {
		panic(fmt.Sprintf("image %q not found but must be found", imageName))
	}

	for _, relatedImageName := range image.dependsOn().relatedImageNameList() {
		if err, errImagesStack := c.validateImageInfiniteLoop(relatedImageName, imageNameStack); err != nil {
			return err, append([]string{imageName}, errImagesStack...)
		}
	}

	return nil, nil
}

func (c *WerfConfig) GroupImagesByIndependentSets(imagesToProcess ImagesToProcess) (sets [][]ImageInterface, err error) {
	if imagesToProcess.WithoutImages {
		return nil, nil
	}

	images := c.getSpecificImages(imagesToProcess)
	sets = [][]ImageInterface{}
	isRelativeChecked := map[ImageInterface]bool{}
	imageRelativesListToHandle := c.getImageRelativesInOrder(images)

	for len(imageRelativesListToHandle) != 0 {
		var currentRelatives []ImageInterface

	outerLoop:
		for image, relatives := range imageRelativesListToHandle {
			for _, relativeImage := range relatives {
				_, ok := isRelativeChecked[relativeImage]
				if !ok {
					continue outerLoop
				}
			}

			currentRelatives = append(currentRelatives, image)
		}

		for _, relativeImage := range currentRelatives {
			isRelativeChecked[relativeImage] = true
			delete(imageRelativesListToHandle, relativeImage)
		}

		sets = append(sets, currentRelatives)
	}

	return sets, nil
}

func (c *WerfConfig) getImageRelativesInOrder(images []ImageInterface) map[ImageInterface][]ImageInterface {
	imageRelatives := map[ImageInterface][]ImageInterface{}
	stack := images

	for len(stack) != 0 {
		current := stack[0]
		stack = stack[1:]

		var relatives []ImageInterface
		depends := c.updateDependencies(current)
		for _, imageName := range depends.relatedImageNameList() {
			relatives = append(relatives, c.GetImage(imageName))
		}
		imageRelatives[current] = relatives

	outerLoop:
		for _, relativeImage := range imageRelatives[current] {
			for key := range imageRelatives {
				if key == relativeImage {
					continue outerLoop
				}
			}

			stack = append(stack, relativeImage)
		}
	}

	return imageRelatives
}

type imageGraph struct {
	ImageName string    `yaml:"image,omitempty"`
	DependsOn DependsOn `yaml:"dependsOn,omitempty"`
}

type DependsOn struct {
	From         string   `yaml:"from,omitempty"`
	Imports      []string `yaml:"import,omitempty"`
	Dependencies []string `yaml:"dependencies,omitempty"`
}

func (d DependsOn) relatedImageNameList() []string {
	var list []string

	if d.From != "" {
		list = append(list, d.From)
	}

	list = append(list, d.Imports...)

	return append(list, d.Dependencies...)
}

func (c *WerfConfig) GetImageGraphList(imagesToProcess ImagesToProcess) ([]imageGraph, error) {
	images := c.getSpecificImages(imagesToProcess)
	var graphList []imageGraph
	for _, image := range images {
		graph := imageGraph{}
		graph.ImageName = image.GetName()
		graph.DependsOn = c.updateDependencies(image)
		graphList = append(graphList, graph)
	}

	return graphList, nil
}

func (c *WerfConfig) updateDependencies(image ImageInterface) DependsOn {
	var d DependsOn
	if !image.IsStapel() {
		return image.dependsOn()
	}
	curDeps := image.dependsOn()

	from := image.GetFrom()
	if c.GetImage(from) != nil {
		d.From = from
	} else {
		image.SetFromExternal()
	}

	d.Dependencies = curDeps.Dependencies
	d.Imports = curDeps.Imports

	return d
}
