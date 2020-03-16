package storage

import (
	"github.com/flant/werf/pkg/image"
)

type ImagesRepo interface {
	GetRepoImage(imageName, tag string) (*image.Info, error)
	GetRepoImages(imageNames []string) (map[string][]*image.Info, error)
	DeleteRepoImage(options DeleteRepoImageOptions, repoImageList ...*image.Info) error

	FetchExistingTags(imageName string) ([]string, error)
	PublishImage(publishImage *image.Image) error

	CreateImageRepo(imageName string) error
	RemoveImageRepo(imageName string) error

	ImageRepositoryName(imageName string) string
	ImageRepositoryNameWithTag(imageName, tag string) string
	ImageRepositoryTag(imageName, tag string) string

	String() string
}

func NewImagesRepo(projectName string, imagesRepoManager *ImagesRepoManager) (ImagesRepo, error) {
	return NewDockerImagesRepo(projectName, imagesRepoManager)
}
