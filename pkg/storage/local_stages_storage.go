package storage

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	v1 "github.com/google/go-containerregistry/pkg/v1"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/container_backend"
	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/util"
)

const (
	LocalStage_ImageRepoFormat         = "%s"
	LocalStage_ImageFormatWithUniqueID = "%s:%s-%d"
	LocalStage_ImageFormat             = "%s:%s"

	LocalImportMetadata_ImageNameFormat = "werf-import-metadata/%s"
	LocalImportMetadata_TagFormat       = "%s"

	LocalClientIDRecord_ImageNameFormat = "werf-client-id/%s"
	LocalClientIDRecord_ImageFormat     = "werf-client-id/%s:%s-%d"

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

func (storage *LocalStagesStorage) FilterStagesAndProcessRelatedData(ctx context.Context, stageDescriptions []*image.StageDescription, opts FilterStagesAndProcessRelatedDataOptions) ([]*image.StageDescription, error) {
	containersOpts := container_backend.ContainersOptions{}
	for _, stageDescription := range stageDescriptions {
		containersOpts.Filters = append(containersOpts.Filters, image.ContainerFilter{Ancestor: stageDescription.Info.ID})
	}
	containers, err := storage.ContainerBackend.Containers(ctx, containersOpts)
	if err != nil {
		return nil, err
	}

	var stageDescriptionListToExcept []*image.StageDescription
	var containerListToRemove []image.Container
	for _, container := range containers {
		for _, stageDescription := range stageDescriptions {
			imageInfo := stageDescription.Info

			if imageInfo.ID == container.ImageID {
				switch {
				case opts.SkipUsedImage:
					logboek.Context(ctx).Default().LogFDetails("Skip image %s (used by container %s)\n", imageInfo.LogName(), container.LogName())
					stageDescriptionListToExcept = append(stageDescriptionListToExcept, stageDescription)
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

	return exceptStageDescriptionList(stageDescriptions, stageDescriptionListToExcept...), nil
}

func exceptStageDescriptionList(stageDescriptionList []*image.StageDescription, stageDescriptionListToExcept ...*image.StageDescription) []*image.StageDescription {
	var result []*image.StageDescription

loop:
	for _, sd1 := range stageDescriptionList {
		for _, sd2 := range stageDescriptionListToExcept {
			if sd2 == sd1 {
				continue loop
			}
		}

		result = append(result, sd1)
	}

	return result
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
	imagesOpts.Filters = append(imagesOpts.Filters, util.NewPair("label", fmt.Sprintf("%s=%s", image.WerfCacheVersionLabel, image.BuildCacheVersion)))

	images, err := storage.ContainerBackend.Images(ctx, imagesOpts)
	if err != nil {
		return nil, fmt.Errorf("unable to list images: %w", err)
	}
	return images.ConvertToStages()
}

func (storage *LocalStagesStorage) GetStagesIDsByDigest(ctx context.Context, projectName, digest string, opts ...Option) ([]image.StageID, error) {
	imagesOpts := container_backend.ImagesOptions{}
	imagesOpts.Filters = append(imagesOpts.Filters, util.NewPair("reference", fmt.Sprintf(LocalStage_ImageRepoFormat, projectName)))
	// NOTE digest already depends on build-cache-version
	imagesOpts.Filters = append(imagesOpts.Filters, util.NewPair("label", fmt.Sprintf("%s=%s", image.WerfStageDigestLabel, digest)))

	images, err := storage.ContainerBackend.Images(ctx, imagesOpts)
	if err != nil {
		return nil, fmt.Errorf("unable to get docker images: %w", err)
	}
	return images.ConvertToStages()
}

func (storage *LocalStagesStorage) GetStageDescription(ctx context.Context, projectName string, stageID image.StageID) (*image.StageDescription, error) {
	stageImageName := storage.ConstructStageImageName(projectName, stageID.Digest, stageID.UniqueID)
	info, err := storage.ContainerBackend.GetImageInfo(ctx, stageImageName, container_backend.GetImageInfoOpts{})
	if err != nil {
		return nil, fmt.Errorf("unable to get image %s info: %w", stageImageName, err)
	}

	if info != nil {
		return &image.StageDescription{
			StageID: image.NewStageID(stageID.Digest, stageID.UniqueID),
			Info:    info,
		}, nil
	}
	return nil, nil
}

func (storage *LocalStagesStorage) ExportStage(ctx context.Context, stageDescription *image.StageDescription, destinationReference string, mutateConfigFunc func(config v1.Config) (v1.Config, error)) error {
	if err := storage.ContainerBackend.Tag(ctx, stageDescription.Info.Name, destinationReference, container_backend.TagOpts{}); err != nil {
		return fmt.Errorf("unable to tag %q as %q: %w", stageDescription.Info.Name, destinationReference, err)
	}
	defer func() {
		_ = storage.ContainerBackend.Rmi(ctx, destinationReference, container_backend.RmiOpts{Force: true})
	}()

	if err := storage.ContainerBackend.Push(ctx, destinationReference, container_backend.PushOpts{}); err != nil {
		return fmt.Errorf("unable to push %q: %w", destinationReference, err)
	}
	return docker_registry.API().MutateAndPushImage(ctx, destinationReference, destinationReference, mutateExportStageConfig(mutateConfigFunc))
}

func (storage *LocalStagesStorage) DeleteStage(ctx context.Context, stageDescription *image.StageDescription, options DeleteImageOptions) error {
	var imageReferences []string
	imageInfo := stageDescription.Info

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

func (storage *LocalStagesStorage) AddStageCustomTag(ctx context.Context, stageDescription *image.StageDescription, tag string) error {
	return fmt.Errorf("not implemented")
}

func (storage *LocalStagesStorage) CheckStageCustomTag(ctx context.Context, stageDescription *image.StageDescription, tag string) error {
	return fmt.Errorf("not implemented")
}

func (storage *LocalStagesStorage) DeleteStageCustomTag(ctx context.Context, tag string) error {
	return fmt.Errorf("not implemented")
}

func (storage *LocalStagesStorage) RejectStage(ctx context.Context, projectName, digest string, uniqueID int64) error {
	return nil
}

func (storage *LocalStagesStorage) ConstructStageImageName(projectName, digest string, uniqueID int64) string {
	if uniqueID == 0 {
		return fmt.Sprintf(LocalStage_ImageFormat, projectName, digest)
	}
	return fmt.Sprintf(LocalStage_ImageFormatWithUniqueID, projectName, digest, uniqueID)
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

func (storage *LocalStagesStorage) GetClientIDRecords(ctx context.Context, projectName string, opts ...Option) ([]*ClientIDRecord, error) {
	logboek.Context(ctx).Debug().LogF("-- LocalStagesStorage.GetClientID for project %s\n", projectName)

	imagesOpts := container_backend.ImagesOptions{}
	imagesOpts.Filters = append(imagesOpts.Filters, util.NewPair("reference", fmt.Sprintf(LocalClientIDRecord_ImageNameFormat, projectName)))
	images, err := storage.ContainerBackend.Images(ctx, imagesOpts)
	if err != nil {
		return nil, fmt.Errorf("unable to list images: %w", err)
	}

	var res []*ClientIDRecord
	for _, img := range images {
		for _, repoTag := range img.RepoTags {
			_, tag := image.ParseRepositoryAndTag(repoTag)

			tagParts := strings.SplitN(util.Reverse(tag), "-", 2)
			if len(tagParts) != 2 {
				continue
			}

			clientID, timestampMillisecStr := util.Reverse(tagParts[1]), util.Reverse(tagParts[0])

			timestampMillisec, err := strconv.ParseInt(timestampMillisecStr, 10, 64)
			if err != nil {
				continue
			}

			rec := &ClientIDRecord{ClientID: clientID, TimestampMillisec: timestampMillisec}
			res = append(res, rec)

			logboek.Context(ctx).Debug().LogF("-- LocalStagesStorage.GetClientID got clientID record: %s\n", rec)
		}
	}
	return res, nil
}

func (storage *LocalStagesStorage) PostClientIDRecord(ctx context.Context, projectName string, rec *ClientIDRecord) error {
	logboek.Context(ctx).Debug().LogF("-- LocalStagesStorage.PostClientID %s for project %s\n", rec.ClientID, projectName)

	fullImageName := fmt.Sprintf(LocalClientIDRecord_ImageFormat, projectName, rec.ClientID, rec.TimestampMillisec)
	labels := []string{fmt.Sprintf("%s=%s", image.WerfLabel, projectName)}

	logboek.Context(ctx).Debug().LogF("-- LocalStagesStorage.PostClientID full image name: %s\n", fullImageName)

	if err := storage.ContainerBackend.PostManifest(ctx, fullImageName, container_backend.PostManifestOpts{Labels: labels}); err != nil {
		return fmt.Errorf("unable to post %q: %w", fullImageName, err)
	}

	logboek.Context(ctx).Info().LogF("Posted new clientID %q for project %s\n", rec.ClientID, projectName)

	return nil
}

func (storage *LocalStagesStorage) PostMultiplatformImage(ctx context.Context, projectName, tag string, allPlatformsImages []*image.Info) error {
	return nil
}

func (storage *LocalStagesStorage) String() string {
	return LocalStorageAddress
}

func (storage *LocalStagesStorage) Address() string {
	return LocalStorageAddress
}

func (storage *LocalStagesStorage) GetStageCustomTagMetadataIDs(ctx context.Context, opts ...Option) ([]string, error) {
	return nil, nil
}

func (storage *LocalStagesStorage) GetStageCustomTagMetadata(ctx context.Context, tagOrID string) (*CustomTagMetadata, error) {
	return nil, fmt.Errorf("not implemented")
}

func (storage *LocalStagesStorage) RegisterStageCustomTag(ctx context.Context, projectName string, stageDescription *image.StageDescription, tag string) error {
	return nil
}

func (storage *LocalStagesStorage) UnregisterStageCustomTag(ctx context.Context, tag string) error {
	return nil
}

func (storage *LocalStagesStorage) CopyFromStorage(ctx context.Context, src StagesStorage, projectName string, stageID image.StageID, opts CopyFromStorageOptions) (*image.StageDescription, error) {
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
