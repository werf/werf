package utils

import (
	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/storage"
)

func NewImagesRepo(imagesRepoAddress, imageRepoMode, implementationName string, dockerRegistryOptions docker_registry.DockerRegistryOptions) storage.ImagesRepo {
	projectName := ProjectName()

	i, err := storage.NewImagesRepo(
		projectName,
		imagesRepoAddress,
		imageRepoMode,
		storage.ImagesRepoOptions{
			DockerImagesRepoOptions: storage.DockerImagesRepoOptions{
				DockerRegistryOptions: dockerRegistryOptions,
				Implementation:        implementationName,
			},
		},
	)
	Ω(err).ShouldNot(HaveOccurred())

	return i
}

func NewStagesStorage(stagesStorageAddress string, implementationName string, dockerRegistryOptions docker_registry.DockerRegistryOptions) storage.StagesStorage {
	s, err := storage.NewStagesStorage(
		stagesStorageAddress,
		&container_runtime.LocalDockerServerRuntime{},
		storage.StagesStorageOptions{
			RepoStagesStorageOptions: storage.RepoStagesStorageOptions{
				DockerRegistryOptions: dockerRegistryOptions,
				Implementation:        implementationName,
			},
		},
	)
	Ω(err).ShouldNot(HaveOccurred())

	return s
}

func ImagesRepoAllImageRepoTags(imagesRepo storage.ImagesRepo, imageName string) []string {
	tags, err := imagesRepo.GetAllImageRepoTags(imageName)
	Ω(err).ShouldNot(HaveOccurred())
	return tags
}

func StagesStorageRepoImagesCount(stagesStorage storage.StagesStorage) int {
	repoImages, err := stagesStorage.GetAllStages(ProjectName())
	Ω(err).ShouldNot(HaveOccurred())
	return len(repoImages)
}

func StagesStorageManagedImagesCount(stagesStorage storage.StagesStorage) int {
	managedImages, err := stagesStorage.GetManagedImages(ProjectName())
	Ω(err).ShouldNot(HaveOccurred())
	return len(managedImages)
}
