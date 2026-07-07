package utils

import (
	"context"

	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/werf"
)

func NewRegistryStorage(ctx context.Context, stagesStorageAddress, implementationName string, dockerRegistryOptions docker_registry.DockerRegistryOptions) storage.RegistryStorage {
	if stagesStorageAddress == storage.LocalStorageAddress {
		return storage.NewLocalRegistryStorage(container_backend.NewDockerServerBackend(werf.HostLocker().Locker()))
	} else {
		dockerRegistry, err := docker_registry.NewDockerRegistry(ctx, stagesStorageAddress, implementationName, dockerRegistryOptions)
		Expect(err).ShouldNot(HaveOccurred())
		return storage.NewRepoRegistryStorage(&storage.NewRepoRegistryStorageOptions{
			RepoAddress:      stagesStorageAddress,
			ContainerBackend: container_backend.NewDockerServerBackend(werf.HostLocker().Locker()),
			DockerRegistry:   dockerRegistry,
		})
	}
}

func StagesCount(ctx context.Context, registryStorage storage.RegistryStorage) int {
	repoImages, err := registryStorage.GetStagesIDs(WithDependencies(ctx), ProjectName())
	Expect(err).ShouldNot(HaveOccurred())
	return len(repoImages)
}

func ManagedImagesCount(ctx context.Context, registryStorage storage.RegistryStorage) int {
	managedImages, err := registryStorage.GetManagedImages(WithDependencies(ctx), ProjectName())
	Expect(err).ShouldNot(HaveOccurred())
	return len(managedImages)
}

func CustomTagsMetadataList(ctx context.Context, registryStorage storage.RegistryStorage) []*storage.CustomTagMetadata {
	ctx = WithDependencies(ctx)

	customTagMetadataIDs, err := registryStorage.GetStageCustomTagMetadataIDs(ctx)
	Expect(err).ShouldNot(HaveOccurred())

	var result []*storage.CustomTagMetadata
	for _, metadataID := range customTagMetadataIDs {
		customTagMetadata, err := registryStorage.GetStageCustomTagMetadata(ctx, metadataID)
		Expect(err).ShouldNot(HaveOccurred())
		result = append(result, customTagMetadata)
	}

	return result
}

func ImageMetadata(ctx context.Context, registryStorage storage.RegistryStorage, imageName string) map[string][]string {
	imageMetadataByImageName, _, err := registryStorage.GetAllAndGroupImageMetadataByImageName(WithDependencies(ctx), ProjectName(), []string{imageName})
	Expect(err).ShouldNot(HaveOccurred())
	return imageMetadataByImageName[imageName]
}
