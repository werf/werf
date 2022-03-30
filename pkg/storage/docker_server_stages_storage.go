package storage

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/container_backend"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/util"
)

const (
	LocalStage_ImageRepoFormat = "%s"
	LocalStage_ImageFormat     = "%s:%s-%d"

	LocalImportMetadata_ImageNameFormat = "werf-import-metadata/%s"
	LocalImportMetadata_TagFormat       = "%s"

	LocalClientIDRecord_ImageNameFormat = "werf-client-id/%s"
	LocalClientIDRecord_ImageFormat     = "werf-client-id/%s:%s-%d"
)

const ImageDeletionFailedDueToUsedByContainerErrorTip = "Use --force option to remove all containers that are based on deleting werf docker images"

func IsImageDeletionFailedDueToUsingByContainerErr(err error) bool {
	return strings.HasSuffix(err.Error(), ImageDeletionFailedDueToUsedByContainerErrorTip)
}

func getDigestAndUniqueIDFromLocalStageImageTag(repoStageImageTag string) (string, int64, error) {
	parts := strings.SplitN(repoStageImageTag, "-", 2)
	if uniqueID, err := image.ParseUniqueIDAsTimestamp(parts[1]); err != nil {
		return "", 0, fmt.Errorf("unable to parse unique id %s as timestamp: %s", parts[1], err)
	} else {
		return parts[0], uniqueID, nil
	}
}

type DockerServerStagesStorage struct {
	// Local stages storage is compatible only with docker-server backed runtime
	DockerServerBackend *container_backend.DockerServerBackend
}

func NewDockerServerStagesStorage(dockerServerBackend *container_backend.DockerServerBackend) *DockerServerStagesStorage {
	return &DockerServerStagesStorage{DockerServerBackend: dockerServerBackend}
}

func (storage *DockerServerStagesStorage) ConstructStageImageName(projectName, digest string, uniqueID int64) string {
	return fmt.Sprintf(LocalStage_ImageFormat, projectName, digest, uniqueID)
}

func (storage *DockerServerStagesStorage) GetStagesIDs(ctx context.Context, projectName string, _ ...Option) ([]image.StageID, error) {
	filterSet := localStagesStorageFilterSetBase(projectName)
	images, err := docker.Images(ctx, types.ImageListOptions{Filters: filterSet})
	if err != nil {
		return nil, fmt.Errorf("unable to get docker images: %s", err)
	}

	return convertToStagesList(images)
}

func (storage *DockerServerStagesStorage) ExportStage(ctx context.Context, stageDescription *image.StageDescription, destinationReference string) error {
	if err := docker.CliTag(ctx, stageDescription.Info.Name, destinationReference); err != nil {
		return err
	}
	defer func() { _ = docker.CliRmi(ctx, destinationReference) }()

	if err := docker.CliPushWithRetries(ctx, destinationReference); err != nil {
		return err
	}

	return docker_registry.API().MutateAndPushImage(ctx, destinationReference, destinationReference, mutateExportStageConfig)
}

func (storage *DockerServerStagesStorage) DeleteStage(ctx context.Context, stageDescription *image.StageDescription, options DeleteImageOptions) error {
	return deleteRepoImageListInDockerServerStagesStorage(ctx, stageDescription, options.RmiForce)
}

func (storage *DockerServerStagesStorage) RejectStage(_ context.Context, _, _ string, _ int64) error {
	return nil
}

type FilterStagesAndProcessRelatedDataOptions struct {
	SkipUsedImage            bool
	RmForce                  bool
	RmContainersThatUseImage bool
}

func (storage *DockerServerStagesStorage) FilterStagesAndProcessRelatedData(ctx context.Context, stageDescriptions []*image.StageDescription, options FilterStagesAndProcessRelatedDataOptions) ([]*image.StageDescription, error) {
	return processRelatedContainers(ctx, stageDescriptions, processRelatedContainersOptions{
		skipUsedImages:           options.SkipUsedImage,
		rmContainersThatUseImage: options.RmContainersThatUseImage,
		rmForce:                  options.RmForce,
	})
}

func (storage *DockerServerStagesStorage) CreateRepo(_ context.Context) error {
	return nil
}

func (storage *DockerServerStagesStorage) DeleteRepo(_ context.Context) error {
	return nil
}

func (storage *DockerServerStagesStorage) GetStageDescription(ctx context.Context, projectName, digest string, uniqueID int64) (*image.StageDescription, error) {
	stageImageName := storage.ConstructStageImageName(projectName, digest, uniqueID)

	inspect, err := storage.DockerServerBackend.GetImageInspect(ctx, stageImageName)
	if err != nil {
		return nil, fmt.Errorf("unable to get image %s inspect: %s", stageImageName, err)
	}

	if inspect != nil {
		return &image.StageDescription{
			StageID: &image.StageID{Digest: digest, UniqueID: uniqueID},
			Info:    image.NewInfoFromInspect(stageImageName, inspect),
		}, nil
	}

	return nil, nil
}

func (storage *DockerServerStagesStorage) AddStageCustomTag(_ context.Context, _ *image.StageDescription, _ string) error {
	return fmt.Errorf("not implemented")
}

func (storage *DockerServerStagesStorage) CheckStageCustomTag(_ context.Context, _ *image.StageDescription, _ string) error {
	return fmt.Errorf("not implemented")
}

func (storage *DockerServerStagesStorage) DeleteStageCustomTag(_ context.Context, _ string) error {
	return fmt.Errorf("not implemented")
}

func (storage *DockerServerStagesStorage) GetStageCustomTagMetadata(_ context.Context, _ string) (*CustomTagMetadata, error) {
	return nil, fmt.Errorf("not implemented")
}

func (storage *DockerServerStagesStorage) GetStageCustomTagMetadataIDs(_ context.Context, _ ...Option) ([]string, error) {
	return nil, nil
}

func (storage *DockerServerStagesStorage) AddManagedImage(_ context.Context, _, _ string) error {
	return nil
}

func (storage *DockerServerStagesStorage) RmManagedImage(_ context.Context, _, _ string) error {
	return nil
}

func (storage *DockerServerStagesStorage) IsManagedImageExist(_ context.Context, _, _ string, _ ...Option) (bool, error) {
	return false, nil
}

func (storage *DockerServerStagesStorage) GetManagedImages(_ context.Context, _ string, _ ...Option) ([]string, error) {
	return []string{}, nil
}

func (storage *DockerServerStagesStorage) GetStagesIDsByDigest(ctx context.Context, projectName, digest string, _ ...Option) ([]image.StageID, error) {
	filterSet := filters.NewArgs()
	filterSet.Add("reference", fmt.Sprintf(LocalStage_ImageRepoFormat, projectName))
	// NOTE digest already depends on build-cache-version
	filterSet.Add("label", fmt.Sprintf("%s=%s", image.WerfStageDigestLabel, digest))

	images, err := docker.Images(ctx, types.ImageListOptions{Filters: filterSet})
	if err != nil {
		return nil, fmt.Errorf("unable to get docker images: %s", err)
	}

	return convertToStagesList(images)
}

func (storage *DockerServerStagesStorage) ShouldFetchImage(_ context.Context, _ container_backend.LegacyImageInterface) (bool, error) {
	return false, nil
}

func (storage *DockerServerStagesStorage) FetchImage(_ context.Context, _ container_backend.LegacyImageInterface) error {
	return nil
}

func (storage *DockerServerStagesStorage) StoreImage(ctx context.Context, img container_backend.LegacyImageInterface) error {
	return storage.DockerServerBackend.TagImageByName(ctx, img)
}

func (storage *DockerServerStagesStorage) PutImageMetadata(_ context.Context, _, _, _, _ string) error {
	return nil
}

func (storage *DockerServerStagesStorage) RmImageMetadata(_ context.Context, _, _, _, _ string) error {
	return nil
}

func (storage *DockerServerStagesStorage) selectFullImageMetadataName(_ context.Context, _, _, _, _ string) (string, error) {
	return "", nil
}

func (storage *DockerServerStagesStorage) IsImageMetadataExist(_ context.Context, _, _, _, _ string, _ ...Option) (bool, error) {
	return false, nil
}

func (storage *DockerServerStagesStorage) GetAllAndGroupImageMetadataByImageName(_ context.Context, _ string, _ []string, _ ...Option) (map[string]map[string][]string, map[string]map[string][]string, error) {
	return map[string]map[string][]string{}, map[string]map[string][]string{}, nil
}

func (storage *DockerServerStagesStorage) GetImportMetadata(ctx context.Context, projectName, id string) (*ImportMetadata, error) {
	logboek.Context(ctx).Debug().LogF("-- DockerServerStagesStorage.GetImportMetadata %s %s\n", projectName, id)

	fullImageName := makeLocalImportMetadataName(projectName, id)
	logboek.Context(ctx).Debug().LogF("-- DockerServerStagesStorage.GetImportMetadata full image name: %s\n", fullImageName)

	inspect, err := storage.DockerServerBackend.GetImageInspect(ctx, fullImageName)
	if err != nil {
		return nil, fmt.Errorf("unable to get image %s inspect: %s", fullImageName, err)
	}

	if inspect != nil {
		return newImportMetadataFromLabels(inspect.Config.Labels), nil
	}

	return nil, nil
}

func (storage *DockerServerStagesStorage) PutImportMetadata(ctx context.Context, projectName string, metadata *ImportMetadata) error {
	logboek.Context(ctx).Debug().LogF("-- DockerServerStagesStorage.PutImportMetadata %s %v\n", projectName, metadata)

	fullImageName := makeLocalImportMetadataName(projectName, metadata.ImportSourceID)
	logboek.Context(ctx).Debug().LogF("-- DockerServerStagesStorage.PutImportMetadata full image name: %s\n", fullImageName)

	if exists, err := docker.ImageExist(ctx, fullImageName); err != nil {
		return fmt.Errorf("unable to check existence of image %q: %s", fullImageName, err)
	} else if exists {
		return nil
	}

	labels := metadata.ToLabels()
	labels[image.WerfLabel] = projectName

	if err := docker.CreateImage(ctx, fullImageName, labels); err != nil {
		return fmt.Errorf("unable to create image %q: %s", fullImageName, err)
	}

	return nil
}

func (storage *DockerServerStagesStorage) RmImportMetadata(ctx context.Context, projectName, id string) error {
	logboek.Context(ctx).Debug().LogF("-- DockerServerStagesStorage.RmImportMetadata %s %s\n", projectName, id)

	fullImageName := makeLocalImportMetadataName(projectName, id)
	logboek.Context(ctx).Debug().LogF("-- DockerServerStagesStorage.RmImportMetadata full image name: %s\n", fullImageName)

	if exists, err := docker.ImageExist(ctx, fullImageName); err != nil {
		return fmt.Errorf("unable to check existence of image %s: %s", fullImageName, err)
	} else if !exists {
		return nil
	}

	if err := docker.CliRmi(ctx, "--force", fullImageName); err != nil {
		return fmt.Errorf("unable to remove image %s: %s", fullImageName, err)
	}

	return nil
}

func (storage *DockerServerStagesStorage) GetImportMetadataIDs(ctx context.Context, projectName string, _ ...Option) ([]string, error) {
	logboek.Context(ctx).Debug().LogF("-- DockerServerStagesStorage.GetImportMetadataIDs %s\n", projectName)

	filterSet := filters.NewArgs()
	filterSet.Add("reference", fmt.Sprintf(LocalImportMetadata_ImageNameFormat, projectName))

	images, err := docker.Images(ctx, types.ImageListOptions{Filters: filterSet})
	if err != nil {
		return nil, fmt.Errorf("unable to get docker images: %s", err)
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

func makeLocalImportMetadataName(projectName, importSourceID string) string {
	return strings.Join(
		[]string{
			fmt.Sprintf(LocalImportMetadata_ImageNameFormat, projectName),
			fmt.Sprintf(LocalImportMetadata_TagFormat, importSourceID),
		}, ":",
	)
}

func (storage *DockerServerStagesStorage) String() string {
	return LocalStorageAddress
}

func (storage *DockerServerStagesStorage) Address() string {
	return LocalStorageAddress
}

func (storage *DockerServerStagesStorage) GetClientIDRecords(ctx context.Context, projectName string, _ ...Option) ([]*ClientIDRecord, error) {
	logboek.Context(ctx).Debug().LogF("-- DockerServerStagesStorage.GetClientID for project %s\n", projectName)

	filterSet := filters.NewArgs()
	filterSet.Add("reference", fmt.Sprintf(LocalClientIDRecord_ImageNameFormat, projectName))

	images, err := docker.Images(ctx, types.ImageListOptions{Filters: filterSet})
	if err != nil {
		return nil, fmt.Errorf("unable to get docker images: %s", err)
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

			logboek.Context(ctx).Debug().LogF("-- DockerServerStagesStorage.GetClientID got clientID record: %s\n", rec)
		}
	}

	return res, nil
}

func (storage *DockerServerStagesStorage) PostClientIDRecord(ctx context.Context, projectName string, rec *ClientIDRecord) error {
	logboek.Context(ctx).Debug().LogF("-- DockerServerStagesStorage.PostClientID %s for project %s\n", rec.ClientID, projectName)

	fullImageName := fmt.Sprintf(LocalClientIDRecord_ImageFormat, projectName, rec.ClientID, rec.TimestampMillisec)

	logboek.Context(ctx).Debug().LogF("-- DockerServerStagesStorage.PostClientID full image name: %s\n", fullImageName)

	labels := map[string]string{image.WerfLabel: projectName}
	if err := docker.CreateImage(ctx, fullImageName, labels); err != nil {
		return fmt.Errorf("unable to create image %q: %s", fullImageName, err)
	}

	logboek.Context(ctx).Info().LogF("Posted new clientID %q for project %s\n", rec.ClientID, projectName)

	return nil
}

type processRelatedContainersOptions struct {
	skipUsedImages           bool
	rmContainersThatUseImage bool
	rmForce                  bool
}

func processRelatedContainers(ctx context.Context, stageDescriptionList []*image.StageDescription, options processRelatedContainersOptions) ([]*image.StageDescription, error) {
	filterSet := filters.NewArgs()
	for _, stageDescription := range stageDescriptionList {
		filterSet.Add("ancestor", stageDescription.Info.ID)
	}

	containerList, err := containerListByFilterSet(ctx, filterSet)
	if err != nil {
		return nil, err
	}

	var stageDescriptionListToExcept []*image.StageDescription
	var containerListToRemove []types.Container
	for _, container := range containerList {
		for _, stageDescription := range stageDescriptionList {
			imageInfo := stageDescription.Info

			if imageInfo.ID == container.ImageID {
				switch {
				case options.skipUsedImages:
					logboek.Context(ctx).Default().LogFDetails("Skip image %s (used by container %s)\n", logImageName(imageInfo), logContainerName(container))
					stageDescriptionListToExcept = append(stageDescriptionListToExcept, stageDescription)
				case options.rmContainersThatUseImage:
					containerListToRemove = append(containerListToRemove, container)
				default:
					return nil, fmt.Errorf("cannot remove image %s used by container %s\n%s", logImageName(imageInfo), logContainerName(container), ImageDeletionFailedDueToUsedByContainerErrorTip)
				}
			}
		}
	}

	if err := deleteContainers(ctx, containerListToRemove, options.rmForce); err != nil {
		return nil, err
	}

	return exceptStageDescriptionList(stageDescriptionList, stageDescriptionListToExcept...), nil
}

func containerListByFilterSet(ctx context.Context, filterSet filters.Args) ([]types.Container, error) {
	containersOptions := types.ContainerListOptions{}
	containersOptions.All = true
	containersOptions.Quiet = true
	containersOptions.Filters = filterSet

	return docker.Containers(ctx, containersOptions)
}

func deleteContainers(ctx context.Context, containers []types.Container, rmForce bool) error {
	for _, container := range containers {
		if err := docker.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{Force: rmForce}); err != nil {
			return err
		}
	}

	return nil
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

func localStagesStorageFilterSetBase(projectName string) filters.Args {
	filterSet := filters.NewArgs()
	filterSet.Add("reference", fmt.Sprintf(LocalStage_ImageRepoFormat, projectName))
	filterSet.Add("label", fmt.Sprintf("%s=%s", image.WerfLabel, projectName))
	filterSet.Add("label", fmt.Sprintf("%s=%s", image.WerfCacheVersionLabel, image.BuildCacheVersion))
	return filterSet
}

func logImageName(image *image.Info) string {
	if image.Name == "<none>:<none>" {
		return image.ID
	} else {
		return image.Name
	}
}

func logContainerName(container types.Container) string {
	name := container.ID
	if len(container.Names) != 0 {
		name = container.Names[0]
	}

	return name
}

func convertToStagesList(imageSummaryList []types.ImageSummary) ([]image.StageID, error) {
	var stagesList []image.StageID

	for _, imageSummary := range imageSummaryList {
		repoTags := imageSummary.RepoTags
		if len(repoTags) == 0 {
			repoTags = append(repoTags, "<none>:<none>")
		}

		for _, repoTag := range repoTags {
			_, tag := image.ParseRepositoryAndTag(repoTag)

			if len(tag) != 70 || len(strings.Split(tag, "-")) != 2 { // 2604b86b2c7a1c6d19c62601aadb19e7d5c6bb8f17bc2bf26a390ea7-1611836746968
				continue
			}

			if digest, uniqueID, err := getDigestAndUniqueIDFromLocalStageImageTag(tag); err != nil {
				return nil, err
			} else {
				stagesList = append(stagesList, image.StageID{Digest: digest, UniqueID: uniqueID})
			}
		}
	}

	return stagesList, nil
}

func deleteRepoImageListInDockerServerStagesStorage(ctx context.Context, stageDescription *image.StageDescription, rmiForce bool) error {
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

	if err := imageReferencesRemove(ctx, imageReferences, rmiForce); err != nil {
		return err
	}

	return nil
}

func imageReferencesRemove(ctx context.Context, references []string, rmiForce bool) error {
	if len(references) == 0 {
		return nil
	}

	var args []string
	if rmiForce {
		args = append(args, "--force")
	}
	args = append(args, references...)

	if err := docker.CliRmi_LiveOutput(ctx, args...); err != nil {
		return err
	}

	return nil
}
