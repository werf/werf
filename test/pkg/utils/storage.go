package utils

import (
	"context"

	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/werf"
)

func NewStagesStorage(ctx context.Context, stagesStorageAddress, implementationName string, dockerRegistryOptions docker_registry.DockerRegistryOptions) storage.CacheAndMetaStorage {
	dockerRegistry, err := docker_registry.NewDockerRegistry(ctx, stagesStorageAddress, implementationName, dockerRegistryOptions)
	Expect(err).ShouldNot(HaveOccurred())
	return storage.NewRepoStagesStorage(&storage.NewRepoStagesStorageOptions{
		RepoAddress:      stagesStorageAddress,
		ContainerBackend: container_backend.NewDockerServerBackend(werf.HostLocker().Locker()),
		DockerRegistry:   dockerRegistry,
	})
}

func StagesCount(ctx context.Context, stagesStorage storage.StageReader) int {
	repoImages, err := stagesStorage.GetStagesIDs(WithDependencies(ctx), ProjectName())
	Expect(err).ShouldNot(HaveOccurred())
	return len(repoImages)
}

func ManagedImagesCount(ctx context.Context, stagesStorage storage.MetaStorage) int {
	managedImages, err := stagesStorage.GetManagedImages(WithDependencies(ctx), ProjectName())
	Expect(err).ShouldNot(HaveOccurred())
	return len(managedImages)
}

func CustomTagsMetadataList(ctx context.Context, stagesStorage storage.ImagesRepoStorage) []*storage.CustomTagMetadata {
	ctx = WithDependencies(ctx)

	customTagMetadataIDs, err := stagesStorage.GetStageCustomTagMetadataIDs(ctx)
	Expect(err).ShouldNot(HaveOccurred())

	var result []*storage.CustomTagMetadata
	for _, metadataID := range customTagMetadataIDs {
		customTagMetadata, err := stagesStorage.GetStageCustomTagMetadata(ctx, metadataID)
		Expect(err).ShouldNot(HaveOccurred())
		result = append(result, customTagMetadata)
	}

	return result
}

func ImageMetadata(ctx context.Context, stagesStorage storage.MetaStorage, imageName string) map[string][]string {
	imageMetadataByImageName, _, err := stagesStorage.GetAllAndGroupImageMetadataByImageName(WithDependencies(ctx), ProjectName(), []string{imageName})
	Expect(err).ShouldNot(HaveOccurred())
	return imageMetadataByImageName[imageName]
}
