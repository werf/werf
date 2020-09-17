package utils

import (
	"context"

	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/storage"
)

func NewImagesRepo(ctx context.Context, imagesRepoAddress, imageRepoMode, implementationName string, dockerRegistryOptions docker_registry.DockerRegistryOptions) storage.ImagesRepo {
	projectName := ProjectName()

	i, err := storage.NewImagesRepo(
		ctx,
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

func ImagesRepoAllImageRepoTags(ctx context.Context, imagesRepo storage.ImagesRepo, imageName string) []string {
	tags, err := imagesRepo.GetAllImageRepoTags(ctx, imageName)
	Ω(err).ShouldNot(HaveOccurred())
	return tags
}

func StagesStorageRepoImagesCount(ctx context.Context, stagesStorage storage.StagesStorage) int {
	repoImages, err := stagesStorage.GetStagesIDs(ctx, ProjectName())
	Ω(err).ShouldNot(HaveOccurred())
	return len(repoImages)
}

func StagesStorageManagedImagesCount(ctx context.Context, stagesStorage storage.StagesStorage) int {
	managedImages, err := stagesStorage.GetManagedImages(ctx, ProjectName())
	Ω(err).ShouldNot(HaveOccurred())
	return len(managedImages)
}
