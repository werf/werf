package config

import (
	"errors"
	"fmt"
	"strings"
)

type WerfConfig struct {
	Meta   *Meta
	images []ImageInterface
}

func (c *WerfConfig) CheckImagesExistence(imageNameList []string, onlyFinal bool) error {
	for _, name := range imageNameList {
		image := c.GetImage(name)
		if image == nil {
			return fmt.Errorf("image %q not defined in werf.yaml", name)
		}

		if onlyFinal && !image.IsFinal() {
			return fmt.Errorf("image %q is not final", name)
		}
	}

	return nil
}

func (c *WerfConfig) GetSpecificImages(imageNameList []string, onlyFinal bool) ([]ImageInterface, error) {
	if len(imageNameList) == 0 {
		return c.Images(onlyFinal), nil
	}

	if err := c.CheckImagesExistence(imageNameList, onlyFinal); err != nil {
		return nil, err
	}

	var resultImages []ImageInterface
	for _, name := range imageNameList {
		image := c.GetImage(name)
		if image == nil {
			panic(fmt.Sprintf("image %q not found but must be found", name))
		}

		if onlyFinal && !image.IsFinal() {
			continue
		}

		resultImages = append(resultImages, image)
	}

	return resultImages, nil
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
	for _, image := range c.Images(false) {
		if image.GetName() == imageName {
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

func (c *WerfConfig) GroupImagesByIndependentSets(imageNameList []string) (sets [][]ImageInterface, err error) {
	images, err := c.GetSpecificImages(imageNameList, true)
	if err != nil {
		return nil, err
	}

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
		for _, imageName := range current.dependsOn().relatedImageNameList() {
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

func (c *WerfConfig) GetImageGraphList(imageNameList []string) ([]imageGraph, error) {
	images, err := c.GetSpecificImages(imageNameList, true)
	if err != nil {
		return nil, err
	}

	var graphList []imageGraph
	for _, image := range images {
		graph := imageGraph{}
		graph.ImageName = image.GetName()
		graph.DependsOn = image.dependsOn()
		graphList = append(graphList, graph)
	}

	return graphList, nil
}
