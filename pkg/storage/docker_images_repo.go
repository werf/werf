package storage

import (
	"fmt"

	"github.com/flant/werf/pkg/docker_registry"
	"github.com/flant/werf/pkg/image"
)

type DockerImagesRepo struct {
	docker_registry.DockerRegistry
	*ImagesRepoManager // FIXME rename images repo manager to something
	projectName        string
}

func NewDockerImagesRepo(projectName string, imagesRepoManager *ImagesRepoManager, dockerRegistryOptions docker_registry.APIOptions) (ImagesRepo, error) {
	dockerRegistry, err := docker_registry.NewDockerRegistry(imagesRepoManager.ImagesRepo(), dockerRegistryOptions)
	if err != nil {
		return nil, err
	}

	imagesRepo := &DockerImagesRepo{
		projectName:       projectName,
		ImagesRepoManager: imagesRepoManager,
		DockerRegistry:    dockerRegistry,
	}

	return imagesRepo, nil
}

func (repo *DockerImagesRepo) GetRepoImage(imageName, tag string) (*image.Info, error) {
	return repo.DockerRegistry.GetRepoImage(repo.ImageRepositoryNameWithTag(imageName, tag))
}

func (repo *DockerImagesRepo) GetRepoImages(imageNames []string) (map[string][]*image.Info, error) {
	if repo.ImagesRepoManager.IsMonorepo() {
		return repo.getRepoImagesFromMonorepo(imageNames)
	} else {
		return repo.getRepoImagesFromMultirepo(imageNames)
	}
}

func (repo *DockerImagesRepo) DeleteRepoImage(_ DeleteRepoImageOptions, repoImageList ...*image.Info) error {
	return repo.DockerRegistry.DeleteRepoImage(repoImageList...)
}

func (repo *DockerImagesRepo) CreateImageRepo(_ string) error {
	return nil
}

func (repo *DockerImagesRepo) RemoveImageRepo(_ string) error {
	return nil
}

func (repo *DockerImagesRepo) FetchExistingTags(imageName string) ([]string, error) {
	imageRepoName := repo.ImagesRepoManager.ImageRepo(imageName)
	if existingTags, err := repo.DockerRegistry.Tags(imageRepoName); err != nil {
		return nil, fmt.Errorf("unable to get docker tags for image %q: %s", imageRepoName, err)
	} else {
		return existingTags, nil
	}
}

// FIXME: use docker-registry object
func (repo *DockerImagesRepo) PublishImage(publishImage *image.Image) error {
	return publishImage.Export()
}

func (repo *DockerImagesRepo) ImageRepositoryName(imageName string) string {
	return repo.ImagesRepoManager.ImageRepo(imageName)
}

func (repo *DockerImagesRepo) ImageRepositoryNameWithTag(imageName, tag string) string {
	return repo.ImagesRepoManager.ImageRepoWithTag(imageName, tag)
}

func (repo *DockerImagesRepo) ImageRepositoryTag(imageName, tag string) string {
	return repo.ImagesRepoManager.ImageRepoTag(imageName, tag)
}

func (repo *DockerImagesRepo) String() string {
	return repo.ImagesRepoManager.ImagesRepo()
}

func (repo *DockerImagesRepo) getRepoImagesFromMonorepo(imageNames []string) (map[string][]*image.Info, error) {
	tags, err := repo.selectImages(repo.ImagesRepoManager.imagesRepo)
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

func (repo *DockerImagesRepo) getRepoImagesFromMultirepo(imageNames []string) (map[string][]*image.Info, error) {
	imageTags := map[string][]*image.Info{}
	for _, imageName := range imageNames {
		tags, err := repo.selectImages(repo.ImagesRepoManager.ImageRepo(imageName))
		if err != nil {
			return nil, err
		}

		imageTags[imageName] = tags
	}

	return imageTags, nil
}

func (repo *DockerImagesRepo) selectImages(reference string) ([]*image.Info, error) {
	return repo.DockerRegistry.SelectRepoImageList(reference, func(info *image.Info) bool {
		werfImageLabel, ok := info.Labels[image.WerfImageLabel]
		return ok && werfImageLabel == "true"
	})
}
