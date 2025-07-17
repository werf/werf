package utils

import (
	"context"

	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/storage"
)

func NewStagesStorage(ctx context.Context, stagesStorageAddress, implementationName string, dockerRegistryOptions docker_registry.DockerRegistryOptions) storage.PrimaryStagesStorage {
	if stagesStorageAddress == storage.LocalStorageAddress {
		return storage.NewLocalStagesStorage(container_backend.NewDockerServerBackend())
	} else {
		dockerRegistry, err := docker_registry.NewDockerRegistry(ctx, stagesStorageAddress, implementationName, dockerRegistryOptions)
		Expect(err).ShouldNot(HaveOccurred())
		return storage.NewRepoStagesStorage(stagesStorageAddress, &container_backend.DockerServerBackend{}, dockerRegistry)
	}
}

func StagesCount(ctx context.Context, stagesStorage storage.StagesStorage) int {
	repoImages, err := stagesStorage.GetStagesIDs(WithDependencies(ctx), ProjectName())
	Expect(err).ShouldNot(HaveOccurred())
	return len(repoImages)
}

func ManagedImagesCount(ctx context.Context, stagesStorage storage.StagesStorage) int {
	managedImages, err := stagesStorage.GetManagedImages(WithDependencies(ctx), ProjectName())
	Expect(err).ShouldNot(HaveOccurred())
	return len(managedImages)
}

func CustomTagsMetadataList(ctx context.Context, stagesStorage storage.PrimaryStagesStorage) []*storage.CustomTagMetadata {
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

func ImageMetadata(ctx context.Context, stagesStorage storage.StagesStorage, imageName string) map[string][]string {
	imageMetadataByImageName, _, err := stagesStorage.GetAllAndGroupImageMetadataByImageName(WithDependencies(ctx), ProjectName(), []string{imageName})
	Expect(err).ShouldNot(HaveOccurred())
	return imageMetadataByImageName[imageName]
}

func ImportMetadataIDs(ctx context.Context, stagesStorage storage.StagesStorage) []string {
	ids, err := stagesStorage.GetImportMetadataIDs(WithDependencies(ctx), ProjectName())
	Expect(err).ShouldNot(HaveOccurred())
	return ids
}

func RmImportMetadata(ctx context.Context, stagesStorage storage.StagesStorage, importSourceID string) {
	err := stagesStorage.RmImportMetadata(WithDependencies(ctx), ProjectName(), importSourceID)
	Expect(err).ShouldNot(HaveOccurred())
}
