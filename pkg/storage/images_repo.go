package storage

import (
	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/image"
)

type ImagesRepo interface {
	GetRepoImage(imageName, tag string) (*image.Info, error)
	GetRepoImages(imageNames []string) (map[string][]*image.Info, error)
	SelectRepoImages(imageNames []string, f func(string, *image.Info, error) (bool, error)) (map[string][]*image.Info, error)
	DeleteRepoImage(_ DeleteImageOptions, repoImageList ...*image.Info) error

	GetAllImageRepoTags(imageName string) ([]string, error)
	PublishImage(publishImage *container_runtime.WerfImage) error

	CreateImageRepo(imageName string) error
	DeleteImageRepo(imageName string) error

	ImageRepositoryName(imageName string) string
	ImageRepositoryNameWithTag(imageName, tag string) string
	ImageRepositoryTag(imageName, tag string) string

	String() string
}

type ImagesRepoOptions struct {
	DockerImagesRepoOptions
}

func NewImagesRepo(projectName, imagesRepoAddress, imagesRepoMode string, options ImagesRepoOptions) (ImagesRepo, error) {
	return NewDockerImagesRepo(projectName, imagesRepoAddress, imagesRepoMode, options.DockerImagesRepoOptions)
}
