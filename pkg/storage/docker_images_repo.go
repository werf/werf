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
	return repo.DockerRegistry.GetRepoImage(ctx, repo.ImageRepositoryNameWithTag(imageName, tag))
}

func (repo *DockerImagesRepo) GetRepoImages(ctx context.Context, imageNames []string) (map[string][]*image.Info, error) {
	return repo.SelectRepoImages(ctx, imageNames, nil)
}

func (repo *DockerImagesRepo) SelectRepoImages(ctx context.Context, imageNames []string, f func(string, *image.Info, error) (bool, error)) (map[string][]*image.Info, error) {
	if repo.imagesRepoManager.IsMonorepo() {
		return repo.getRepoImagesFromMonorepo(ctx, imageNames, f)
	} else {
		return repo.getRepoImagesFromMultirepo(ctx, imageNames, f)
	}
}

func (repo *DockerImagesRepo) DeleteRepoImage(ctx context.Context, _ DeleteImageOptions, repoImageList ...*image.Info) error {
	return repo.DockerRegistry.DeleteRepoImage(ctx, repoImageList...)
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

func (repo *DockerImagesRepo) ImageRepositoryNameWithTag(imageName, tag string) string {
	return repo.imagesRepoManager.ImageRepoWithTag(imageName, tag)
}

func (repo *DockerImagesRepo) ImageRepositoryTag(imageName, tag string) string {
	return repo.imagesRepoManager.ImageRepoTag(imageName, tag)
}

func (repo *DockerImagesRepo) String() string {
	return repo.imagesRepoManager.ImagesRepo()
}

func (repo *DockerImagesRepo) getRepoImagesFromMonorepo(ctx context.Context, imageNames []string, f func(string, *image.Info, error) (bool, error)) (map[string][]*image.Info, error) {
	tags, err := repo.selectImages(ctx, repo.imagesRepoManager.imagesRepo, f)
	if err != nil {
		return nil, err
	}

	imageTags := map[string][]*image.Info{}

loop:
	for _, info := range tags {
		for _, imageName := range imageNames {
			metaImageName, ok := info.Labels[image.WerfImageNameLabel]
			if !ok {
				continue
			}

			if metaImageName == imageName {
				imageTags[imageName] = append(imageTags[imageName], info)
				continue loop
			}
		}
	}

	return imageTags, nil
}

func (repo *DockerImagesRepo) getRepoImagesFromMultirepo(ctx context.Context, imageNames []string, f func(string, *image.Info, error) (bool, error)) (map[string][]*image.Info, error) {
	imageTags := map[string][]*image.Info{}
	for _, imageName := range imageNames {
		tags, err := repo.selectImages(ctx, repo.imagesRepoManager.ImageRepo(imageName), f)
		if err != nil {
			return nil, err
		}

		imageTags[imageName] = tags
	}

	return imageTags, nil
}

func (repo *DockerImagesRepo) selectImages(ctx context.Context, reference string, f func(string, *image.Info, error) (bool, error)) ([]*image.Info, error) {
	return repo.DockerRegistry.SelectRepoImageList(ctx, reference, func(ref string, info *image.Info, err error) (bool, error) {
		if err != nil {
			if f != nil {
				return f(ref, info, err)
			}

			return false, err
		}

		werfImageLabel, ok := info.Labels[image.WerfImageLabel]
		if !ok || werfImageLabel != "true" {
			return false, nil
		}

		if f != nil {
			ok, err := f(ref, info, err)
			if err != nil {
				return false, err
			}

			if !ok {
				return false, nil
			}
		}

		return true, nil
	})
}
