package storage

import (
	"context"

	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/image"
)

type ImagesRepo interface {
	GetRepoImage(ctx context.Context, imageName, tag string) (*image.Info, error)
	GetRepoImages(ctx context.Context, imageNames []string) (map[string][]*image.Info, error)
	SelectRepoImages(ctx context.Context, imageNames []string, f func(string, *image.Info, error) (bool, error)) (map[string][]*image.Info, error)
	DeleteRepoImage(ctx context.Context, repoImageList ...*image.Info) error

	GetAllImageRepoTags(ctx context.Context, imageName string) ([]string, error)
	PublishImage(ctx context.Context, publishImage *container_runtime.WerfImage) error

	CreateImageRepo(ctx context.Context, imageName string) error
	DeleteImageRepo(ctx context.Context, imageName string) error

	ImageRepositoryName(imageName string) string
	ImageRepositoryNameWithTag(imageName, tag string) string
	ImageRepositoryTag(imageName, tag string) string

	String() string
}

type ImagesRepoOptions struct {
	DockerImagesRepoOptions
}

func NewImagesRepo(ctx context.Context, projectName, imagesRepoAddress, imagesRepoMode string, options ImagesRepoOptions) (ImagesRepo, error) {
	return NewDockerImagesRepo(ctx, projectName, imagesRepoAddress, imagesRepoMode, options.DockerImagesRepoOptions)
}
