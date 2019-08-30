package common

import (
	"fmt"
	"strings"
)

const (
	MultirepoImagesRepoMode   = "multirepo"
	MonorepoImagesRepoMode    = "monorepo"
	MonorepoTagPartsSeparator = "-"
)

type ImagesRepoManager struct {
	imagesRepo            string
	namelessImageRepoFunc func(imagesRepo string) string
	imageRepoFunc         func(imagesRepo, imageName string) string
	imageRepoTagFunc      func(imageName, tag string) string
}

func newImagesRepoManager(
	imagesRepo string,
	namelessImageRepoFunc func(imagesRepo string) string,
	imageRepoFunc func(imagesRepo, imageName string) string,
	imageRepoTagFunc func(imageName, tag string) string) *ImagesRepoManager {

	formattedImagesRepo := strings.TrimRight(imagesRepo, "/")

	return &ImagesRepoManager{
		imagesRepo:            formattedImagesRepo,
		namelessImageRepoFunc: namelessImageRepoFunc,
		imageRepoFunc:         imageRepoFunc,
		imageRepoTagFunc:      imageRepoTagFunc,
	}
}

func (m *ImagesRepoManager) ImagesRepo() string {
	return m.imagesRepo
}

func (m *ImagesRepoManager) ImageRepo(imageName string) string {
	var repo string
	if imageName == "" {
		repo = m.namelessImageRepoFunc(m.imagesRepo)
	} else {
		repo = m.imageRepoFunc(m.imagesRepo, imageName)
	}

	return repo
}

func (m *ImagesRepoManager) ImageRepoTag(imageName, tag string) string {
	return m.imageRepoTagFunc(imageName, tag)
}

func (m *ImagesRepoManager) ImageRepoWithTag(imageName, tag string) string {
	return strings.Join([]string{m.ImageRepo(imageName), m.ImageRepoTag(imageName, tag)}, ":")
}

func (m *ImagesRepoManager) IsMonorepo() bool {
	return m.ImagesRepo() == m.ImageRepo("image")
}

func GetImagesRepoManager(imagesRepo, imagesRepoMode string) (*ImagesRepoManager, error) {
	var namelessImageRepoFunc func(imagesRepo string) string
	var imageRepoFunc func(imagesRepo, imageName string) string
	var imageRepoTagFunc func(imageName, tag string) string

	switch imagesRepoMode {
	case MultirepoImagesRepoMode:
		namelessImageRepoFunc = func(imagesRepo string) string {
			return imagesRepo
		}

		imageRepoFunc = func(imagesRepo, imageName string) string {
			return strings.Join([]string{imagesRepo, imageName}, "/")
		}

		imageRepoTagFunc = func(_, tag string) string {
			return tag
		}
	case MonorepoImagesRepoMode:
		namelessImageRepoFunc = func(imagesRepo string) string {
			return imagesRepo
		}

		imageRepoFunc = func(imagesRepo, _ string) string {
			return imagesRepo
		}

		imageRepoTagFunc = func(imageName, tag string) string {
			if imageName != "" {
				tag = strings.Join([]string{imageName, tag}, MonorepoTagPartsSeparator)
			}

			return tag
		}
	default:
		return nil, fmt.Errorf("bad images repo mode '%s': only %s and %s supported", imagesRepoMode, MultirepoImagesRepoMode, MonorepoImagesRepoMode)
	}

	return newImagesRepoManager(
		imagesRepo,
		namelessImageRepoFunc,
		imageRepoFunc,
		imageRepoTagFunc,
	), nil
}
