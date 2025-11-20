package storage

import (
	"context"
	"fmt"
	"strings"

	v1 "github.com/google/go-containerregistry/pkg/v1"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/docker_registry/api"
	"github.com/werf/werf/v2/pkg/image"
)

const (
	LocalStage_ImageRepoFormat              = "%s"
	LocalStage_ImageFormatWithCreationTs    = "%s:%s-%d"
	FilterReferenceLocalStageByDigestFormat = "%s:%s*"
	LocalStage_ImageFormat                  = "%s:%s"

	LocalImportMetadata_ImageNameFormat = "werf-import-metadata/%s"
	LocalImportMetadata_TagFormat       = "%s"

	ImageDeletionFailedDueToUsedByContainerErrorTip = "Use --force option to remove all containers that are based on deleting werf docker images"
)

func IsImageDeletionFailedDueToUsingByContainerErr(err error) bool {
	return strings.HasSuffix(err.Error(), ImageDeletionFailedDueToUsedByContainerErrorTip)
}

type LocalStagesStorage struct {
	ContainerBackend container_backend.ContainerBackend
}

func NewLocalStagesStorage(containerBackend container_backend.ContainerBackend) *LocalStagesStorage {
	return &LocalStagesStorage{ContainerBackend: containerBackend}
}

func (storage *LocalStagesStorage) FilterStageDescSetAndProcessRelatedData(ctx context.Context, stageDescSet image.StageDescSet, opts FilterStagesAndProcessRelatedDataOptions) (image.StageDescSet, error) {
	containersOpts := container_backend.ContainersOptions{}
	for stageDesc := range stageDescSet.Iter() {
		containersOpts.Filters = append(containersOpts.Filters, image.ContainerFilter{Ancestor: stageDesc.Info.ID})
	}
	containers, err := storage.ContainerBackend.Containers(ctx, containersOpts)
	if err != nil {
		return nil, err
	}

	stageDescSetToExclude := image.NewStageDescSet()
	var containerListToRemove []image.Container
	for _, container := range containers {
		for stageDesc := range stageDescSet.Iter() {
			imageInfo := stageDesc.Info

			if imageInfo.ID == container.ImageID {
				switch {
				case opts.SkipUsedImage:
					logboek.Context(ctx).Default().LogFDetails("Skip image %s (used by container %s)\n", imageInfo.LogName(), container.LogName())
					stageDescSetToExclude.Add(stageDesc)
				case opts.RmContainersThatUseImage:
					containerListToRemove = append(containerListToRemove, container)
				default:
					return nil, fmt.Errorf("cannot remove image %s used by container %s\n%s", imageInfo.LogName(), container.LogName(), ImageDeletionFailedDueToUsedByContainerErrorTip)
				}
			}
		}
	}

	if err := storage.deleteContainers(ctx, containerListToRemove, opts.RmForce); err != nil {
		return nil, err
	}

	return stageDescSet.Difference(stageDescSetToExclude), nil
}

func (storage *LocalStagesStorage) deleteContainers(ctx context.Context, containers []image.Container, rmForce bool) error {
	for _, container := range containers {
		if err := storage.ContainerBackend.Rm(ctx, container.ID, container_backend.RmOpts{Force: rmForce}); err != nil {
			return fmt.Errorf("unable to remove container %q: %w", container.ID, err)
		}
	}
	return nil
}

func (storage *LocalStagesStorage) GetStagesIDs(ctx context.Context, projectName string, opts ...Option) ([]image.StageID, error) {
	imagesOpts := container_backend.ImagesOptions{}
	imagesOpts.Filters = append(imagesOpts.Filters, util.NewPair("reference", fmt.Sprintf(LocalStage_ImageRepoFormat, projectName)))
	imagesOpts.Filters = append(imagesOpts.Filters, util.NewPair("label", fmt.Sprintf("%s=%s", image.WerfLabel, projectName)))

	images, err := storage.ContainerBackend.Images(ctx, imagesOpts)
	if err != nil {
		return nil, fmt.Errorf("unable to list images: %w", err)
	}
	return images.ConvertToStages()
}

func (storage *LocalStagesStorage) GetStagesIDsByDigest(ctx context.Context, projectName, digest string, parentStageCreationTs int64, _ ...Option) ([]image.StageID, error) {
	imagesOpts := container_backend.ImagesOptions{}
	imagesOpts.Filters = append(imagesOpts.Filters, util.NewPair("reference", fmt.Sprintf(FilterReferenceLocalStageByDigestFormat, projectName, digest)))

	images, err := storage.ContainerBackend.Images(ctx, imagesOpts)
	if err != nil {
		return nil, fmt.Errorf("unable to get docker images: %w", err)
	}

	stagesIDs, err := images.ConvertToStages()
	if err != nil {
		return nil, fmt.Errorf("unable to convert images to stages: %w", err)
	}

	var resultStageIDs []image.StageID
	for _, stageID := range stagesIDs {
		if parentStageCreationTs > stageID.CreationTs {
			logboek.Context(ctx).Debug().LogF("Skip stage %s (parent stage creation timestamp %d is greater than the stage creation timestamp %d)\n", stageID.String(), parentStageCreationTs, stageID.CreationTs)
			continue
		}

		resultStageIDs = append(resultStageIDs, stageID)
	}

	return resultStageIDs, nil
}

func (storage *LocalStagesStorage) GetStageDesc(ctx context.Context, projectName string, stageID image.StageID) (*image.StageDesc, error) {
	stageImageName := storage.ConstructStageImageName(projectName, stageID.Digest, stageID.CreationTs)
	info, err := storage.ContainerBackend.GetImageInfo(ctx, stageImageName, container_backend.GetImageInfoOpts{})
	if err != nil {
		return nil, fmt.Errorf("unable to get image %s info: %w", stageImageName, err)
	}

	if info != nil {
		return &image.StageDesc{
			StageID: image.NewStageID(stageID.Digest, stageID.CreationTs),
			Info:    info,
		}, nil
	}
	return nil, nil
}

func (storage *LocalStagesStorage) ExportStage(ctx context.Context, stageDesc *image.StageDesc, destinationReference string, mutateConfigFunc func(config v1.Config) (v1.Config, error)) error {
	if err := storage.ContainerBackend.Tag(ctx, stageDesc.Info.Name, destinationReference, container_backend.TagOpts{}); err != nil {
		return fmt.Errorf("unable to tag %q as %q: %w", stageDesc.Info.Name, destinationReference, err)
	}
	defer func() {
		_ = storage.ContainerBackend.Rmi(ctx, destinationReference, container_backend.RmiOpts{Force: true})
	}()

	if err := storage.ContainerBackend.Push(ctx, destinationReference, container_backend.PushOpts{}); err != nil {
		return fmt.Errorf("unable to push %q: %w", destinationReference, err)
	}
	return docker_registry.API().MutateAndPushImage(ctx, destinationReference, destinationReference, api.WithConfigMutation(mutateExportStageConfig(mutateConfigFunc)))
}

func (storage *LocalStagesStorage) DeleteStage(ctx context.Context, stageDesc *image.StageDesc, options DeleteImageOptions) error {
	var imageReferences []string
	imageInfo := stageDesc.Info

	if imageInfo.Name == "" {
		imageReferences = append(imageReferences, imageInfo.ID)
	} else {
		isDanglingImage := imageInfo.Name == "<none>:<none>"
		isTaglessImage := !isDanglingImage && imageInfo.Tag == "<none>"

		if isDanglingImage || isTaglessImage {
			imageReferences = append(imageReferences, imageInfo.ID)
		} else {
			imageReferences = append(imageReferences, imageInfo.Name)
		}
	}

	for _, ref := range imageReferences {
		if err := storage.ContainerBackend.Rmi(ctx, ref, container_backend.RmiOpts{Force: options.RmiForce}); err != nil {
			return fmt.Errorf("unable to remove %q: %w", ref, err)
		}
	}
	return nil
}

func (storage *LocalStagesStorage) AddStageCustomTag(_ context.Context, _ *image.StageDesc, _ string) error {
	return fmt.Errorf("not implemented")
}

func (storage *LocalStagesStorage) CheckStageCustomTag(_ context.Context, _ *image.StageDesc, _ string) error {
	return fmt.Errorf("not implemented")
}

func (storage *LocalStagesStorage) DeleteStageCustomTag(_ context.Context, _ string) error {
	return fmt.Errorf("not implemented")
}

func (storage *LocalStagesStorage) RejectStage(_ context.Context, _, _ string, _ int64) error {
	return nil
}

func (storage *LocalStagesStorage) ConstructStageImageName(projectName, digest string, creationTs int64) string {
	if creationTs == 0 {
		return fmt.Sprintf(LocalStage_ImageFormat, projectName, digest)
	}
	return fmt.Sprintf(LocalStage_ImageFormatWithCreationTs, projectName, digest, creationTs)
}

func (storage *LocalStagesStorage) FetchImage(ctx context.Context, img container_backend.LegacyImageInterface) error {
	return nil
}

func (storage *LocalStagesStorage) StoreImage(ctx context.Context, img container_backend.LegacyImageInterface) error {
	return storage.ContainerBackend.TagImageByName(ctx, img)
}

func (storage *LocalStagesStorage) ShouldFetchImage(ctx context.Context, img container_backend.LegacyImageInterface) (bool, error) {
	return false, nil
}

func (storage *LocalStagesStorage) CreateRepo(ctx context.Context) error { return nil }

func (storage *LocalStagesStorage) DeleteRepo(ctx context.Context) error { return nil }

func (storage *LocalStagesStorage) AddManagedImage(ctx context.Context, projectName, imageNameOrManagedImageName string) error {
	return nil
}

func (storage *LocalStagesStorage) RmManagedImage(ctx context.Context, projectName, imageNameOrManagedImageName string) error {
	return nil
}

func (storage *LocalStagesStorage) IsManagedImageExist(ctx context.Context, projectName, imageNameOrManagedImageName string, opts ...Option) (bool, error) {
	return false, nil
}

func (storage *LocalStagesStorage) GetManagedImages(ctx context.Context, projectName string, opts ...Option) ([]string, error) {
	return []string{}, nil
}

func (storage *LocalStagesStorage) PutImageMetadata(ctx context.Context, projectName, imageNameOrManagedImageName, commit, stageID string) error {
	return nil
}

func (storage *LocalStagesStorage) RmImageMetadata(ctx context.Context, projectName, imageNameOrManagedImageNameOrImageMetadataID, commit, stageID string) error {
	return nil
}

func (storage *LocalStagesStorage) IsImageMetadataExist(ctx context.Context, projectName, imageNameOrManagedImageName, commit, stageID string, opts ...Option) (bool, error) {
	return false, nil
}

func (storage *LocalStagesStorage) GetAllAndGroupImageMetadataByImageName(ctx context.Context, projectName string, imageNameOrManagedImageList []string, opts ...Option) (map[string]map[string][]string, map[string]map[string][]string, error) {
	return map[string]map[string][]string{}, map[string]map[string][]string{}, nil
}

func (storage *LocalStagesStorage) GetImportMetadata(ctx context.Context, projectName, id string) (*ImportMetadata, error) {
	logboek.Context(ctx).Debug().LogF("-- LocalStagesStorage.GetImportMetadata %s %s\n", projectName, id)

	fullImageName := makeLocalImportMetadataName(projectName, id)
	logboek.Context(ctx).Debug().LogF("-- LocalStagesStorage.GetImportMetadata full image name: %s\n", fullImageName)

	info, err := storage.ContainerBackend.GetImageInfo(ctx, fullImageName, container_backend.GetImageInfoOpts{})
	if err != nil {
		return nil, fmt.Errorf("unable to get image %s info: %w", fullImageName, err)
	}
	if info == nil {
		return nil, nil
	}
	return newImportMetadataFromLabels(info.Labels), nil
}

func (storage *LocalStagesStorage) PutImportMetadata(ctx context.Context, projectName string, metadata *ImportMetadata) error {
	logboek.Context(ctx).Debug().LogF("-- LocalStagesStorage.PutImportMetadata %s %v\n", projectName, metadata)

	fullImageName := makeLocalImportMetadataName(projectName, metadata.ImportSourceID)
	logboek.Context(ctx).Debug().LogF("-- LocalStagesStorage.PutImportMetadata full image name: %s\n", fullImageName)

	if info, err := storage.ContainerBackend.GetImageInfo(ctx, fullImageName, container_backend.GetImageInfoOpts{}); err != nil {
		return fmt.Errorf("unable to check existence of image %s: %w", fullImageName, err)
	} else if info != nil {
		return nil
	}

	labels := metadata.ToLabels()
	labels = append(labels, fmt.Sprintf("%s=%s", image.WerfLabel, projectName))
	if err := storage.ContainerBackend.PostManifest(ctx, fullImageName, container_backend.PostManifestOpts{Labels: labels}); err != nil {
		return fmt.Errorf("unable to post manifest %q: %w", fullImageName, err)
	}
	return nil
}

func (storage *LocalStagesStorage) RmImportMetadata(ctx context.Context, projectName, id string) error {
	logboek.Context(ctx).Debug().LogF("-- LocalStagesStorage.RmImportMetadata %s %s\n", projectName, id)

	fullImageName := makeLocalImportMetadataName(projectName, id)
	logboek.Context(ctx).Debug().LogF("-- LocalStagesStorage.RmImportMetadata full image name: %s\n", fullImageName)

	if info, err := storage.ContainerBackend.GetImageInfo(ctx, fullImageName, container_backend.GetImageInfoOpts{}); err != nil {
		return fmt.Errorf("unable to check existence of image %s: %w", fullImageName, err)
	} else if info != nil {
		return nil
	}

	if err := storage.ContainerBackend.Rmi(ctx, fullImageName, container_backend.RmiOpts{Force: true}); err != nil {
		return fmt.Errorf("unable to remove image %s: %w", fullImageName, err)
	}
	return nil
}

func (storage *LocalStagesStorage) GetImportMetadataIDs(ctx context.Context, projectName string, opts ...Option) ([]string, error) {
	logboek.Context(ctx).Debug().LogF("-- LocalStagesStorage.GetImportMetadataIDs %s\n", projectName)

	imagesOpts := container_backend.ImagesOptions{}
	imagesOpts.Filters = append(imagesOpts.Filters, util.NewPair("reference", fmt.Sprintf(LocalImportMetadata_ImageNameFormat, projectName)))
	images, err := storage.ContainerBackend.Images(ctx, imagesOpts)
	if err != nil {
		return nil, fmt.Errorf("unable to list images: %w", err)
	}

	var tags []string
	for _, img := range images {
		for _, repoTag := range img.RepoTags {
			_, tag := image.ParseRepositoryAndTag(repoTag)
			tags = append(tags, tag)
		}
	}

	return tags, nil
}

func (storage *LocalStagesStorage) GetClientIDRecords(_ context.Context, _ string, _ ...Option) ([]*ClientIDRecord, error) {
	panic("not implemented")
}

func (storage *LocalStagesStorage) PostClientIDRecord(_ context.Context, _ string, _ *ClientIDRecord) error {
	panic("not implemented")
}

func (storage *LocalStagesStorage) PostMultiplatformImage(_ context.Context, _, _ string, _ []*image.Info, _ []string) error {
	return nil
}

func (storage *LocalStagesStorage) String() string {
	return LocalStorageAddress
}

func (storage *LocalStagesStorage) Address() string {
	return LocalStorageAddress
}

func (storage *LocalStagesStorage) GetStageCustomTagMetadataIDs(_ context.Context, _ ...Option) ([]string, error) {
	return nil, nil
}

func (storage *LocalStagesStorage) GetStageCustomTagMetadata(_ context.Context, _ string) (*CustomTagMetadata, error) {
	return nil, fmt.Errorf("not implemented")
}

func (storage *LocalStagesStorage) RegisterStageCustomTag(_ context.Context, _ string, _ *image.StageDesc, tag string) error {
	return nil
}

func (storage *LocalStagesStorage) UnregisterStageCustomTag(_ context.Context, _ string) error {
	return nil
}

func (storage *LocalStagesStorage) CopyFromStorage(_ context.Context, _ StagesStorage, _ string, _ image.StageID, _ CopyFromStorageOptions) (*image.StageDesc, error) {
	panic("not implemented")
}

func makeLocalImportMetadataName(projectName, importSourceID string) string {
	return strings.Join(
		[]string{
			fmt.Sprintf(LocalImportMetadata_ImageNameFormat, projectName),
			fmt.Sprintf(LocalImportMetadata_TagFormat, importSourceID),
		}, ":",
	)
}

func (storage *LocalStagesStorage) GetSyncServerRecords(ctx context.Context, projectName string, opts ...Option) ([]*SyncServerRecord, error) {
	panic("not implemented")
}

func (storage *LocalStagesStorage) PostSyncServerRecord(ctx context.Context, projectName string, rec *SyncServerRecord) error {
	panic("not implemented")
}

func (storage *LocalStagesStorage) GetLastCleanupRecord(ctx context.Context, projectName string, opts ...Option) (*CleanupRecord, error) {
	panic("not implemented")
}

func (storage *LocalStagesStorage) PostLastCleanupRecord(ctx context.Context, projectName string) error {
	panic("not implemented")
}

func (storage *LocalStagesStorage) MutateAndPushImage(ctx context.Context, src, _ string, newConfig image.SpecConfig, stageImage container_backend.LegacyImageInterface) error {
	newId, err := container_backend.MutateAndPushImage(ctx, src, newConfig, storage.ContainerBackend)
	if err != nil {
		return err
	}
	stageImage.SetBuiltID(buildIdFromDigest(newId))
	if err := storage.ContainerBackend.TagImageByName(ctx, stageImage); err != nil {
		return fmt.Errorf("unable to tag image %q: %w", stageImage.Name(), err)
	}
	return nil
}

func buildIdFromDigest(digest string) string {
	if strings.HasPrefix(digest, "sha256:") {
		return digest[7:]
	}
	return digest
}
