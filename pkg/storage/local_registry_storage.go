package storage

import (
	"context"
	"fmt"
	"strings"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"sigs.k8s.io/yaml"

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

	ImageDeletionFailedDueToUsedByContainerErrorTip = "Use --force option to remove all containers that are based on deleting werf docker images"
)

func IsImageDeletionFailedDueToUsingByContainerErr(err error) bool {
	return strings.HasSuffix(err.Error(), ImageDeletionFailedDueToUsedByContainerErrorTip)
}

type LocalRegistryStorage struct {
	ContainerBackend container_backend.ContainerBackend
}

func NewLocalRegistryStorage(containerBackend container_backend.ContainerBackend) *LocalRegistryStorage {
	return &LocalRegistryStorage{ContainerBackend: containerBackend}
}

func (storage *LocalRegistryStorage) FilterStageDescSetAndProcessRelatedData(ctx context.Context, stageDescSet image.StageDescSet, opts FilterStagesAndProcessRelatedDataOptions) (image.StageDescSet, error) {
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

			if imageInfo.ID != container.ImageID {
				continue
			}

			switch {
			case opts.SkipUsedImage:
				logboek.Context(ctx).Default().LogFDetails("Skip image %s (used by container %s)\n", imageInfo.LogName(), container.LogName())
				stageDescSetToExclude.Add(stageDesc)
			case opts.RmContainersThatUseImage:
				containerListToRemove = append(containerListToRemove, container)
			default:
				return nil, fmt.Errorf("cannot remove image %s used by container %s\n%s", imageInfo.LogName(), container.LogName(), ImageDeletionFailedDueToUsedByContainerErrorTip)
			}

			break
		}
	}

	if err := storage.deleteContainers(ctx, containerListToRemove, opts.RmForce); err != nil {
		return nil, err
	}

	return stageDescSet.Difference(stageDescSetToExclude), nil
}

func (storage *LocalRegistryStorage) deleteContainers(ctx context.Context, containers []image.Container, rmForce bool) error {
	removed := make(map[string]struct{}, len(containers))
	for _, container := range containers {
		if _, ok := removed[container.ID]; ok {
			continue
		}
		removed[container.ID] = struct{}{}

		if err := storage.ContainerBackend.Rm(ctx, container.ID, container_backend.RmOpts{Force: rmForce}); err != nil {
			return fmt.Errorf("unable to remove container %q: %w", container.ID, err)
		}
	}
	return nil
}

func (storage *LocalRegistryStorage) GetStagesIDs(ctx context.Context, projectName string, opts ...Option) ([]image.StageID, error) {
	imagesOpts := container_backend.ImagesOptions{}
	imagesOpts.Filters = append(imagesOpts.Filters, util.NewPair("reference", fmt.Sprintf(LocalStage_ImageRepoFormat, projectName)))
	imagesOpts.Filters = append(imagesOpts.Filters, util.NewPair("label", fmt.Sprintf("%s=%s", image.WerfLabel, projectName)))

	images, err := storage.ContainerBackend.Images(ctx, imagesOpts)
	if err != nil {
		return nil, fmt.Errorf("unable to list images: %w", err)
	}
	return images.ConvertToStages()
}

func (storage *LocalRegistryStorage) GetStagesIDsByDigest(ctx context.Context, projectName, digest string, parentStageCreationTs int64, _ ...Option) ([]image.StageID, error) {
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

func (storage *LocalRegistryStorage) GetStageDesc(ctx context.Context, projectName string, stageID image.StageID) (*image.StageDesc, error) {
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

func (storage *LocalRegistryStorage) ExportStage(ctx context.Context, stageDesc *image.StageDesc, destinationReference string, mutateConfigFunc func(config v1.Config) (v1.Config, error)) error {
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

func (storage *LocalRegistryStorage) DeleteStage(ctx context.Context, stageDesc *image.StageDesc, options DeleteImageOptions) error {
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

func (storage *LocalRegistryStorage) AddStageCustomTag(_ context.Context, _ *image.StageDesc, _ string) error {
	return fmt.Errorf("not implemented")
}

func (storage *LocalRegistryStorage) CheckStageCustomTag(_ context.Context, _ *image.StageDesc, _ string) error {
	return fmt.Errorf("not implemented")
}

func (storage *LocalRegistryStorage) DeleteStageCustomTag(_ context.Context, _ string) error {
	return fmt.Errorf("not implemented")
}

func (storage *LocalRegistryStorage) RejectStage(_ context.Context, _, _ string, _ int64) error {
	return nil
}

func (storage *LocalRegistryStorage) GetRejectedStageIDs(_ context.Context, _ ...Option) ([]image.StageID, error) {
	return nil, nil
}

func (storage *LocalRegistryStorage) DeleteRejectedStageImage(_ context.Context, _ image.StageID, _ DeleteImageOptions) error {
	return nil
}

func (storage *LocalRegistryStorage) DeleteRejectedStageRecord(_ context.Context, _ image.StageID, _ DeleteImageOptions) error {
	return nil
}

func (storage *LocalRegistryStorage) ConstructStageImageName(projectName, digest string, creationTs int64) string {
	if creationTs == 0 {
		return fmt.Sprintf(LocalStage_ImageFormat, projectName, digest)
	}
	return fmt.Sprintf(LocalStage_ImageFormatWithCreationTs, projectName, digest, creationTs)
}

func (storage *LocalRegistryStorage) FetchImage(ctx context.Context, img container_backend.LegacyImageInterface) error {
	return nil
}

func (storage *LocalRegistryStorage) StoreImage(ctx context.Context, img container_backend.LegacyImageInterface) error {
	return storage.ContainerBackend.TagImageByName(ctx, img)
}

func (storage *LocalRegistryStorage) ShouldFetchImage(ctx context.Context, img container_backend.LegacyImageInterface) (bool, error) {
	return false, nil
}

func (storage *LocalRegistryStorage) CreateRepo(ctx context.Context) error { return nil }

func (storage *LocalRegistryStorage) DeleteRepo(ctx context.Context) error { return nil }

func (storage *LocalRegistryStorage) AddManagedImage(ctx context.Context, projectName, imageNameOrManagedImageName string) error {
	return nil
}

func (storage *LocalRegistryStorage) RmManagedImage(ctx context.Context, projectName, imageNameOrManagedImageName string) error {
	return nil
}

func (storage *LocalRegistryStorage) IsManagedImageExist(ctx context.Context, projectName, imageNameOrManagedImageName string, opts ...Option) (bool, error) {
	return false, nil
}

func (storage *LocalRegistryStorage) GetManagedImages(ctx context.Context, projectName string, opts ...Option) ([]string, error) {
	return []string{}, nil
}

func (storage *LocalRegistryStorage) PutImageMetadata(ctx context.Context, projectName, imageNameOrManagedImageName, commit, stageID string) error {
	return nil
}

func (storage *LocalRegistryStorage) RmImageMetadata(ctx context.Context, projectName, imageNameOrManagedImageNameOrImageMetadataID, commit, stageID string) error {
	return nil
}

func (storage *LocalRegistryStorage) IsImageMetadataExist(ctx context.Context, projectName, imageNameOrManagedImageName, commit, stageID string, opts ...Option) (bool, error) {
	return false, nil
}

func (storage *LocalRegistryStorage) GetAllAndGroupImageMetadataByImageName(ctx context.Context, projectName string, imageNameOrManagedImageList []string, opts ...Option) (map[string]map[string][]string, map[string]map[string][]string, error) {
	return map[string]map[string][]string{}, map[string]map[string][]string{}, nil
}

func (storage *LocalRegistryStorage) PostMultiplatformImage(_ context.Context, _, _ string, _ []*image.Info, _ []string) error {
	return nil
}

func (storage *LocalRegistryStorage) String() string {
	return LocalStorageAddress
}

func (storage *LocalRegistryStorage) Address() string {
	return LocalStorageAddress
}

func (storage *LocalRegistryStorage) GetStageCustomTagMetadataIDs(_ context.Context, _ ...Option) ([]string, error) {
	return nil, nil
}

func (storage *LocalRegistryStorage) GetStageCustomTagMetadata(_ context.Context, _ string) (*CustomTagMetadata, error) {
	return nil, fmt.Errorf("not implemented")
}

func (storage *LocalRegistryStorage) RegisterStageCustomTag(_ context.Context, _ string, _ *image.StageDesc, tag string) error {
	return nil
}

func (storage *LocalRegistryStorage) UnregisterStageCustomTag(_ context.Context, _ string) error {
	return nil
}

func (storage *LocalRegistryStorage) CopyFromStorage(_ context.Context, _ RegistryStorage, _ string, _ image.StageID, _ CopyFromStorageOptions) (*image.StageDesc, error) {
	panic("not implemented")
}

func (storage *LocalRegistryStorage) GetLastCleanupRecord(ctx context.Context, projectName string, opts ...Option) (*CleanupRecord, error) {
	panic("not implemented")
}

func (storage *LocalRegistryStorage) PostLastCleanupRecord(ctx context.Context, projectName string) error {
	panic("not implemented")
}

func (storage *LocalRegistryStorage) PostManifest(ctx context.Context, ref string, opts container_backend.PostManifestOpts) error {
	if err := storage.ContainerBackend.PostManifest(ctx, ref, opts); err != nil {
		return fmt.Errorf("unable to post manifest %s: %w", ref, err)
	}

	return nil
}

func (storage *LocalRegistryStorage) MutateAndPushImage(ctx context.Context, src, dest string, newConfig image.SpecConfig, stageImage container_backend.LegacyImageInterface) error {
	if err := logboek.Context(ctx).Debug().LogBlock("-- LocalRegistryStorage.MutateAndPushImage imageSpecConfig").DoError(func() error {
		newConfigData, err := yaml.Marshal(newConfig)
		if err != nil {
			return fmt.Errorf("unable to yaml marshal new config: %w", err)
		}

		logboek.Context(ctx).Debug().LogF(string(newConfigData))
		return nil
	}); err != nil {
		return err
	}

	newId, err := container_backend.MutateAndPushImage(ctx, src, stageImage.GetTargetPlatform(), newConfig, storage.ContainerBackend)
	if err != nil {
		return err
	}

	if err := storage.ContainerBackend.Tag(ctx, newId, dest, container_backend.TagOpts{}); err != nil {
		return fmt.Errorf("unable to tag image %q as %q: %w", newId, dest, err)
	}

	return nil
}
