package storage

import (
	"fmt"

	"github.com/flant/werf/pkg/docker_registry"
	"github.com/flant/werf/pkg/image"
)

type ImagesRepo interface {
	GetRepoImages(imageNames []string) (map[string][]*image.Info, error)
	DeleteRepoImage(options DeleteImageOptions, imageInfo ...*image.Info) error

	FetchExistingTags(imageName string) ([]string, error)
	PublishImage(publishImage *image.Image) error

	CreateImageRepo(imageName string) error
	RemoveImageRepo(imageName string) error

	String() string
	ImageRepositoryName(imageName string) string
	ImageRepositoryNameWithTag(imageName, tag string) string
	ImageRepositoryTag(imageName, tag string) string
}

type DockerImagesRepo struct {
	ImagesRepo // FIXME

	docker_registry.DockerRegistry
	*ImagesRepoManager // FIXME rename images repo manager to something
	projectName        string
}

func NewDockerImagesRepo(projectName string, imagesRepoManager *ImagesRepoManager) *DockerImagesRepo {
	return &DockerImagesRepo{
		projectName:       projectName,
		ImagesRepoManager: imagesRepoManager,
	}
}

func (repo *DockerImagesRepo) String() string {
	return repo.ImagesRepoManager.ImagesRepo()
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

// FIXME: use docker-registry object
func (repo *DockerImagesRepo) FetchExistingTags(imageName string) ([]string, error) {
	fullImageName := repo.ImagesRepoManager.ImageRepo(imageName)
	if existingTags, err := docker_registry.Tags(fullImageName); err != nil {
		return nil, fmt.Errorf("unable to get docker tags for image %q: %s", fullImageName, err)
	} else {
		return existingTags, nil
	}
}

// FIXME: use docker-registry object
func (repo *DockerImagesRepo) PublishImage(publishImage *image.Image) error {
	return publishImage.Export()
}

func (repo *DockerImagesRepo) GetRepoImages(imageNames []string) (map[string][]*image.Info, error) {
	if repo.ImagesRepoManager.IsMonorepo() {
		return repo.GetRepoImagesFromMonorepo(imageNames)
	} else {
		return repo.GetRepoImagesFromMultirepo(imageNames)
	}
}

func (repo *DockerImagesRepo) GetRepoImagesFromMonorepo(imageNames []string) (map[string][]*image.Info, error) {
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

func (repo *DockerImagesRepo) GetRepoImagesFromMultirepo(imageNames []string) (map[string][]*image.Info, error) {
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
	return repo.DockerRegistry.Select(reference, func(info *image.Info) bool {
		werfImageLabel, ok := info.Labels[image.WerfImageLabel]
		if !ok {
			return false
		} else if werfImageLabel != "true" {
			return false
		}

		return true
	})
}

func (repo *DockerImagesRepo) DeleteRepoImage(options DeleteImageOptions, imageInfo ...*image.Info) error {
	return nil
}

func (repo *DockerImagesRepo) CreateImageRepo(_ string) error {
	return nil
}

func (repo *DockerImagesRepo) RemoveImageRepo(_ string) error {
	return nil
}

// TODO: методы связанные только с логикой работы с images-repo
// TODO: работа с низкоуровневым registry через интерфейс docker_registry.DockerRegistry
