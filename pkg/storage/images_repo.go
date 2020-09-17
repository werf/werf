package storage

import (
	"context"
	"strings"

	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/image"
)

const tagNotAssociatedWithImageErrorCode = "TAG_NOT_ASSOCIATED_WITH_IMAGE"

type ImagesRepo interface {
	GetAllImageRepoTags(ctx context.Context, imageName string) ([]string, error)
	GetRepoImage(ctx context.Context, imageName, tag string) (*image.Info, error)
	DeleteRepoImage(ctx context.Context, repoImage *image.Info) error
	IsRepoImage(imageName string, repoImage *image.Info) bool

	PublishImage(ctx context.Context, publishImage *container_runtime.WerfImage) error

	CreateImageRepo(ctx context.Context, imageName string) error
	DeleteImageRepo(ctx context.Context, imageName string) error

	ImageRepositoryName(imageName string) string
	ImageRepositoryNameWithTag(imageName, metaTag string) string
	ImageRepositoryTag(imageName, metaTag string) string
	ImageRepositoryMetaTag(imageName, tag string) string
	IsImageRepositoryTag(imageName, tag string) bool

	String() string
}

type ImagesRepoOptions struct {
	DockerImagesRepoOptions
}

func NewImagesRepo(ctx context.Context, projectName, imagesRepoAddress, imagesRepoMode string, options ImagesRepoOptions) (ImagesRepo, error) {
	return NewDockerImagesRepo(ctx, projectName, imagesRepoAddress, imagesRepoMode, options.DockerImagesRepoOptions)
}

func IsTagNotAssociatedWithImageError(err error) bool {
	return strings.Contains(err.Error(), tagNotAssociatedWithImageErrorCode)
}
