package storage

import (
	"fmt"
	"strings"

	"github.com/flant/werf/pkg/docker_registry"
)

const (
	monorepoTagPartsSeparator = "-"
)

type imagesRepoManager struct {
	imagesRepo            string
	namelessImageRepoFunc func(imagesRepo string) string
	imageRepoFunc         func(imagesRepo, imageName string) string
	imageRepoTagFunc      func(imageName, tag string) string
}

func (m *imagesRepoManager) ImagesRepo() string {
	return m.imagesRepo
}

func (m *imagesRepoManager) ImageRepo(imageName string) string {
	var repo string
	if imageName == "" {
		repo = m.namelessImageRepoFunc(m.imagesRepo)
	} else {
		repo = m.imageRepoFunc(m.imagesRepo, imageName)
	}

	return repo
}

func (m *imagesRepoManager) ImageRepoTag(imageName, tag string) string {
	return m.imageRepoTagFunc(imageName, tag)
}

func (m *imagesRepoManager) ImageRepoWithTag(imageName, tag string) string {
	return strings.Join([]string{m.ImageRepo(imageName), m.ImageRepoTag(imageName, tag)}, ":")
}

func (m *imagesRepoManager) IsMonorepo() bool {
	return m.ImagesRepo() == m.ImageRepo("image")
}

func newImagesRepoManager(imagesRepo, imagesRepoMode string) (*imagesRepoManager, error) {
	var namelessImageRepoFunc func(imagesRepo string) string
	var imageRepoFunc func(imagesRepo, imageName string) string
	var imageRepoTagFunc func(imageName, tag string) string

	switch imagesRepoMode {
	case docker_registry.MultirepoRepoMode:
		namelessImageRepoFunc = func(imagesRepo string) string {
			return imagesRepo
		}

		imageRepoFunc = func(imagesRepo, imageName string) string {
			return strings.Join([]string{imagesRepo, imageName}, "/")
		}

		imageRepoTagFunc = func(_, tag string) string {
			return tag
		}
	case docker_registry.MonorepoRepoMode:
		namelessImageRepoFunc = func(imagesRepo string) string {
			return imagesRepo
		}

		imageRepoFunc = func(imagesRepo, _ string) string {
			return imagesRepo
		}

		imageRepoTagFunc = func(imageName, tag string) string {
			if imageName != "" {
				tag = strings.Join([]string{imageName, tag}, monorepoTagPartsSeparator)
			}

			return tag
		}
	default:
		return nil, fmt.Errorf("bad images repo mode '%s': only %s and %s supported", imagesRepoMode, docker_registry.MultirepoRepoMode, docker_registry.MonorepoRepoMode)
	}

	formattedImagesRepo := strings.TrimRight(imagesRepo, "/")

	imagesRepoManager := &imagesRepoManager{
		imagesRepo:            formattedImagesRepo,
		namelessImageRepoFunc: namelessImageRepoFunc,
		imageRepoFunc:         imageRepoFunc,
		imageRepoTagFunc:      imageRepoTagFunc,
	}

	return imagesRepoManager, nil
}
