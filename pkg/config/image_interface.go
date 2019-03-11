package config

import (
	"errors"
)

type ImageInterface interface {
	ImageBaseConfig() *ImageBase
	IsArtifact() bool

	fromImage() *Image
	fromImageArtifact() *ImageArtifact
	imports() []*Import
}

func imageTree(i ImageInterface) (tree []ImageInterface) {
	if i.fromImage() != nil {
		tree = append(tree, imageTree(i.fromImage())...)
	}

	if i.fromImageArtifact() != nil {
		tree = append(tree, imageTree(i.fromImageArtifact())...)
	}

	for _, imp := range i.imports() {
		if imp.Image != nil {
			tree = append(tree, imageTree(imp.Image)...)
		} else {
			tree = append(tree, imageTree(imp.ImageArtifact)...)
		}
	}

	return append(tree, i)
}

func validateImageInfiniteLoop(i ImageInterface, stack []ImageInterface) (error, []string) {
	imageName := i.ImageBaseConfig().Name

	for _, elm := range stack {
		if elm.ImageBaseConfig().Name == imageName {
			return errors.New("infinite loop detected"), []string{imageName}
		}
	}
	stack = append(stack, i)

	if i.fromImage() != nil {
		if err, errImagesStack := validateImageInfiniteLoop(i.fromImage(), stack); err != nil {
			return err, append([]string{imageName}, errImagesStack...)
		}
	}

	if i.fromImageArtifact() != nil {
		if err, errImagesStack := validateImageInfiniteLoop(i.fromImageArtifact(), stack); err != nil {
			return err, append([]string{imageName}, errImagesStack...)
		}
	}

	for _, imp := range i.imports() {
		if imp.Image != nil {
			if err, errImagesStack := validateImageInfiniteLoop(imp.Image, stack); err != nil {
				return err, append([]string{imageName}, errImagesStack...)
			}
		} else {
			if err, errImagesStack := validateImageInfiniteLoop(imp.ImageArtifact, stack); err != nil {
				return err, append([]string{imageName}, errImagesStack...)
			}
		}
	}

	return nil, nil
}

func relatedImageImages(i ImageInterface) (images []ImageInterface) {
	images = append(images, i)
	if i.fromImage() != nil {
		images = append(images, relatedImageImages(i.fromImage())...)
	}

	if i.fromImageArtifact() != nil {
		images = append(images, relatedImageImages(i.fromImageArtifact())...)
	}

	return
}

func headImage(i ImageInterface) ImageInterface {
	if i.fromImage() != nil {
		return headImage(i.fromImage())
	}

	if i.fromImageArtifact() != nil {
		return headImage(i.fromImageArtifact())
	}

	return i
}
