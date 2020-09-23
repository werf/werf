package utils

import (
	"context"

	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/storage"
)

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

func StagesCount(ctx context.Context, stagesStorage storage.StagesStorage) int {
	repoImages, err := stagesStorage.GetStagesIDs(ctx, ProjectName())
	Ω(err).ShouldNot(HaveOccurred())
	return len(repoImages)
}

func ManagedImagesCount(ctx context.Context, stagesStorage storage.StagesStorage) int {
	managedImages, err := stagesStorage.GetManagedImages(ctx, ProjectName())
	Ω(err).ShouldNot(HaveOccurred())
	return len(managedImages)
}

func ImageMetadata(ctx context.Context, stagesStorage storage.StagesStorage, imageName string) map[string][]string {
	imageMetadataByImageName, _, err := stagesStorage.GetAllAndGroupImageMetadataByImageName(ctx, ProjectName(), []string{imageName})
	Ω(err).ShouldNot(HaveOccurred())
	return imageMetadataByImageName[imageName]
}
