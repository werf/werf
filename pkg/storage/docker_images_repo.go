package storage

import (
	"context"
	"fmt"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/image"
)

type DockerImagesRepo struct {
	docker_registry.DockerRegistry
	*imagesRepoManager // FIXME rename images repo manager to something
	projectName        string
}

type DockerImagesRepoOptions struct {
	docker_registry.DockerRegistryOptions
	Implementation string
}

func NewDockerImagesRepo(ctx context.Context, projectName, imagesRepoAddress, imagesRepoMode string, options DockerImagesRepoOptions) (ImagesRepo, error) {
	resolvedImplementation, err := docker_registry.ResolveImplementation(imagesRepoAddress, options.Implementation)
	if err != nil {
		return nil, err
	}
	logboek.Context(ctx).Info().LogLn("Using images repo docker registry implementation:", resolvedImplementation)

	dockerRegistry, err := docker_registry.NewDockerRegistry(imagesRepoAddress, resolvedImplementation, options.DockerRegistryOptions)
	if err != nil {
		return nil, err
	}

	resolvedImagesRepoMode, err := dockerRegistry.ResolveRepoMode(ctx, imagesRepoAddress, imagesRepoMode)
	if err != nil {
		return nil, err
	}
	logboek.Context(ctx).Info().LogLn("Using images repo mode:", resolvedImagesRepoMode)

	imagesRepoManager, err := newImagesRepoManager(imagesRepoAddress, resolvedImagesRepoMode)
	if err != nil {
		return nil, err
	}

	imagesRepo := &DockerImagesRepo{
		projectName:       projectName,
		imagesRepoManager: imagesRepoManager,
		DockerRegistry:    dockerRegistry,
	}

	return imagesRepo, nil
}

func (repo *DockerImagesRepo) CreateImageRepo(ctx context.Context, imageName string) error {
	return repo.DockerRegistry.CreateRepo(ctx, repo.ImageRepositoryName(imageName))
}

func (repo *DockerImagesRepo) DeleteImageRepo(ctx context.Context, imageName string) error {
	return repo.DockerRegistry.DeleteRepo(ctx, repo.ImageRepositoryName(imageName))
}

func (repo *DockerImagesRepo) GetRepoImage(ctx context.Context, imageName, tag string) (*image.Info, error) {
	repoImage, err := repo.DockerRegistry.GetRepoImage(ctx, repo.ImageRepositoryNameWithTag(imageName, tag))
	if err != nil {
		return nil, err
	}

	if !repo.IsRepoImage(imageName, repoImage) {
		return nil, fmt.Errorf("%s: tag %s is not associated with image %s", tagNotAssociatedWithImageErrorCode, repo.imageRepoTagFunc(imageName, tag), imageName)
	}

	return repoImage, nil
}

func (repo *DockerImagesRepo) IsRepoImage(imageName string, repoImage *image.Info) bool {
	if !repo.isRepoImage(repoImage) {
		return false
	}

	if repo.imagesRepoManager.IsMonorepo() {
		metaImageName, ok := repoImage.Labels[image.WerfImageNameLabel]
		if !ok {
			return false
		}

		return metaImageName == imageName
	} else {
		return true
	}
}

func (repo *DockerImagesRepo) isRepoImage(repoImage *image.Info) bool {
	werfImageLabel, ok := repoImage.Labels[image.WerfImageLabel]
	if !ok || werfImageLabel != "true" {
		return false
	}

	return true
}

func (repo *DockerImagesRepo) DeleteRepoImage(ctx context.Context, repoImage *image.Info) error {
	return repo.DockerRegistry.DeleteRepoImage(ctx, repoImage)
}

func (repo *DockerImagesRepo) GetAllImageRepoTags(ctx context.Context, imageName string) ([]string, error) {
	imageRepoName := repo.imagesRepoManager.ImageRepo(imageName)
	if existingTags, err := repo.DockerRegistry.Tags(ctx, imageRepoName); err != nil {
		return nil, fmt.Errorf("unable to get docker tags for image %q: %s", imageRepoName, err)
	} else {
		return existingTags, nil
	}
}

// FIXME: use docker-registry object
func (repo *DockerImagesRepo) PublishImage(ctx context.Context, publishImage *container_runtime.WerfImage) error {
	return publishImage.Export(ctx)
}

func (repo *DockerImagesRepo) ImageRepositoryName(imageName string) string {
	return repo.imagesRepoManager.ImageRepo(imageName)
}

func (repo *DockerImagesRepo) ImageRepositoryNameWithTag(imageName, metaTag string) string {
	return repo.imagesRepoManager.ImageRepoWithTag(imageName, metaTag)
}

func (repo *DockerImagesRepo) ImageRepositoryTag(imageName, metaTag string) string {
	return repo.imagesRepoManager.ImageRepoTag(imageName, metaTag)
}

func (repo *DockerImagesRepo) ImageRepositoryMetaTag(imageName, tag string) string {
	return repo.imagesRepoManager.ImageRepoMetaTag(imageName, tag)
}

func (repo *DockerImagesRepo) IsImageRepositoryTag(imageName, tag string) bool {
	return repo.imagesRepoManager.isImageRepoTag(imageName, tag)
}

func (repo *DockerImagesRepo) String() string {
	return repo.imagesRepoManager.ImagesRepo()
}
