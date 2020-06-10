package storage

import (
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/image"
)

const (
	LocalStage_ImageRepoPrefix = "werf-stages-storage/"
	LocalStage_ImageRepoFormat = "werf-stages-storage/%s"
	LocalStage_ImageFormat     = "werf-stages-storage/%s:%s-%d"

	LocalManagedImageRecord_ImageNameFormat = "werf-managed-images/%s"
	LocalManagedImageRecord_ImageFormat     = "werf-managed-images/%s:%s"
)

func getSignatureAndUniqueIDFromLocalStageImageTag(repoStageImageTag string) (string, int64, error) {
	parts := strings.SplitN(repoStageImageTag, "-", 2)
	if uniqueID, err := image.ParseUniqueIDAsTimestamp(parts[1]); err != nil {
		return "", 0, fmt.Errorf("unable to parse unique id %s as timestamp: %s", parts[1], err)
	} else {
		return parts[0], uniqueID, nil
	}
}

type LocalDockerServerStagesStorage struct {
	// Local stages storage is compatible only with docker-server backed runtime
	LocalDockerServerRuntime *container_runtime.LocalDockerServerRuntime
}

func NewLocalDockerServerStagesStorage(localDockerServerRuntime *container_runtime.LocalDockerServerRuntime) *LocalDockerServerStagesStorage {
	return &LocalDockerServerStagesStorage{LocalDockerServerRuntime: localDockerServerRuntime}
}

func (storage *LocalDockerServerStagesStorage) ConstructStageImageName(projectName, signature string, uniqueID int64) string {
	return fmt.Sprintf(LocalStage_ImageFormat, projectName, signature, uniqueID)
}

func (storage *LocalDockerServerStagesStorage) GetAllStages(projectName string) ([]image.StageID, error) {
	filterSet := localStagesStorageFilterSetBase(projectName)
	images, err := docker.Images(types.ImageListOptions{Filters: filterSet})
	if err != nil {
		return nil, fmt.Errorf("unable to get docker images: %s", err)
	}

	return convertToStagesList(images)
}

func (storage *LocalDockerServerStagesStorage) DeleteStages(options DeleteImageOptions, stages ...*image.StageDescription) error {
	var imageInfoList []*image.Info
	for _, stageDesc := range stages {
		imageInfoList = append(imageInfoList, stageDesc.Info)
	}

	var err error
	imageInfoList, err = processRelatedContainers(imageInfoList, processRelatedContainersOptions{
		skipUsedImages:           options.SkipUsedImage,
		rmContainersThatUseImage: options.RmContainersThatUseImage,
		rmForce:                  options.RmForce,
	})
	if err != nil {
		return err
	}

	if err := deleteRepoImageListInLocalDockerServerStagesStorage(imageInfoList, options.RmiForce); err != nil {
		return err
	}

	return nil
}

func (storage *LocalDockerServerStagesStorage) CreateRepo() error {
	return nil
}

func (storage *LocalDockerServerStagesStorage) DeleteRepo() error {
	return nil
}

func makeLocalManagedImageRecord(projectName, imageName string) string {
	tag := imageName
	if imageName == "" {
		tag = NamelessImageRecordTag
	}

	tag = strings.ReplaceAll(tag, "/", "__slash__")
	tag = strings.ReplaceAll(tag, "+", "__plus__")

	return fmt.Sprintf(LocalManagedImageRecord_ImageFormat, projectName, tag)
}

func (storage *LocalDockerServerStagesStorage) GetStageDescription(projectName, signature string, uniqueID int64) (*image.StageDescription, error) {
	stageImageName := storage.ConstructStageImageName(projectName, signature, uniqueID)

	if inspect, err := storage.LocalDockerServerRuntime.GetImageInspect(stageImageName); err != nil {
		return nil, fmt.Errorf("unable to get image %s inspect: %s", stageImageName, err)
	} else if inspect != nil {
		return &image.StageDescription{
			StageID: &image.StageID{Signature: signature, UniqueID: uniqueID},
			Info:    image.NewInfoFromInspect(stageImageName, inspect),
		}, nil
	} else {
		return nil, nil
	}
}

func (storage *LocalDockerServerStagesStorage) AddManagedImage(projectName, imageName string) error {
	logboek.Debug.LogF("-- LocalDockerServerStagesStorage.AddManagedImage %s %s\n", projectName, imageName)

	fullImageName := makeLocalManagedImageRecord(projectName, imageName)

	if exsts, err := docker.ImageExist(fullImageName); err != nil {
		return fmt.Errorf("unable to check existence of image %q: %s", fullImageName, err)
	} else if exsts {
		return nil
	}

	if err := docker.CreateImage(fullImageName); err != nil {
		return fmt.Errorf("unable to create image %q: %s", fullImageName, err)
	}
	return nil
}

func (storage *LocalDockerServerStagesStorage) RmManagedImage(projectName, imageName string) error {
	logboek.Debug.LogF("-- LocalDockerServerStagesStorage.RmManagedImage %s %s\n", projectName, imageName)

	fullImageName := makeLocalManagedImageRecord(projectName, imageName)

	if exsts, err := docker.ImageExist(fullImageName); err != nil {
		return fmt.Errorf("unable to check existence of image %q: %s", fullImageName, err)
	} else if !exsts {
		return nil
	}

	if err := docker.CliRmi("--force", fullImageName); err != nil {
		return fmt.Errorf("unable to remove image %q: %s", fullImageName, err)
	}

	return nil
}

func (storage *LocalDockerServerStagesStorage) GetManagedImages(projectName string) ([]string, error) {
	logboek.Debug.LogF("-- LocalDockerServerStagesStorage.GetManagedImages %s\n", projectName)

	filterSet := filters.NewArgs()
	filterSet.Add("reference", fmt.Sprintf(LocalManagedImageRecord_ImageNameFormat, projectName))

	images, err := docker.Images(types.ImageListOptions{Filters: filterSet})
	if err != nil {
		return nil, fmt.Errorf("unable to get docker images: %s", err)
	}

	var res []string
	for _, img := range images {
		for _, repoTag := range img.RepoTags {
			_, tag := image.ParseRepositoryAndTag(repoTag)
			tag = strings.ReplaceAll(tag, "__slash__", "/")
			tag = strings.ReplaceAll(tag, "__plus__", "+")

			if tag == NamelessImageRecordTag {
				res = append(res, "")
			} else {
				res = append(res, tag)
			}
		}
	}
	return res, nil
}

func (storage *LocalDockerServerStagesStorage) GetStagesBySignature(projectName, signature string) ([]image.StageID, error) {
	filterSet := filters.NewArgs()
	filterSet.Add("reference", fmt.Sprintf(LocalStage_ImageRepoFormat, projectName))
	// NOTE signature already depends on build-cache-version
	filterSet.Add("label", fmt.Sprintf("%s=%s", image.WerfStageSignatureLabel, signature))

	images, err := docker.Images(types.ImageListOptions{Filters: filterSet})
	if err != nil {
		return nil, fmt.Errorf("unable to get docker images: %s", err)
	}

	return convertToStagesList(images)
}

func (storage *LocalDockerServerStagesStorage) ShouldFetchImage(_ container_runtime.Image) (bool, error) {
	return false, nil
}

func (storage *LocalDockerServerStagesStorage) FetchImage(_ container_runtime.Image) error {
	return nil
}

func (storage *LocalDockerServerStagesStorage) StoreImage(img container_runtime.Image) error {
	return storage.LocalDockerServerRuntime.TagBuiltImageByName(img)
}

func (storage *LocalDockerServerStagesStorage) String() string {
	return LocalStorageAddress
}

func (storage *LocalDockerServerStagesStorage) Address() string {
	return LocalStorageAddress
}

type processRelatedContainersOptions struct {
	skipUsedImages           bool
	rmContainersThatUseImage bool
	rmForce                  bool
}

func processRelatedContainers(imageInfoList []*image.Info, options processRelatedContainersOptions) ([]*image.Info, error) {
	filterSet := filters.NewArgs()
	for _, imgInfo := range imageInfoList {
		filterSet.Add("ancestor", imgInfo.ID)
	}

	containerList, err := containerListByFilterSet(filterSet)
	if err != nil {
		return nil, err
	}

	var imageInfoListToExcept []*image.Info
	var containerListToRemove []types.Container
	for _, container := range containerList {
		for _, imgInfo := range imageInfoList {
			if imgInfo.ID == container.ImageID {
				if options.skipUsedImages {
					logboek.Default.LogFDetails("Skip image %s (used by container %s)\n", logImageName(imgInfo), logContainerName(container))
					imageInfoListToExcept = append(imageInfoListToExcept, imgInfo)
				} else if options.rmContainersThatUseImage {
					containerListToRemove = append(containerListToRemove, container)
				} else {
					return nil, fmt.Errorf("cannot remove image %s used by container %s\n%s", logImageName(imgInfo), logContainerName(container), "Use --force option to remove all containers that are based on deleting werf docker images")
				}
			}
		}
	}

	if err := deleteContainers(containerListToRemove, options.rmForce); err != nil {
		return nil, err
	}

	return exceptRepoImageList(imageInfoList, imageInfoListToExcept...), nil
}

func containerListByFilterSet(filterSet filters.Args) ([]types.Container, error) {
	containersOptions := types.ContainerListOptions{}
	containersOptions.All = true
	containersOptions.Quiet = true
	containersOptions.Filters = filterSet

	return docker.Containers(containersOptions)
}

func deleteContainers(containers []types.Container, rmForce bool) error {
	for _, container := range containers {
		if err := docker.ContainerRemove(container.ID, types.ContainerRemoveOptions{Force: rmForce}); err != nil {
			return err
		}
	}

	return nil
}

func exceptRepoImageList(imageInfoList []*image.Info, imageInfoListToExcept ...*image.Info) []*image.Info {
	var newImageInfoList []*image.Info

loop:
	for _, imgInfo := range imageInfoList {
		for _, repoImageToExcept := range imageInfoListToExcept {
			if repoImageToExcept == imgInfo {
				continue loop
			}
		}

		newImageInfoList = append(newImageInfoList, imgInfo)
	}

	return newImageInfoList
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
			if signature, uniqueID, err := getSignatureAndUniqueIDFromLocalStageImageTag(tag); err != nil {
				return nil, err
			} else {
				stagesList = append(stagesList, image.StageID{Signature: signature, UniqueID: uniqueID})
			}
		}
	}

	return stagesList, nil
}

func deleteRepoImageListInLocalDockerServerStagesStorage(imageInfoList []*image.Info, rmiForce bool) error {
	var imageReferences []string
	for _, imgInfo := range imageInfoList {
		if imgInfo.Name == "" {
			imageReferences = append(imageReferences, imgInfo.ID)
		} else {
			isDanglingImage := imgInfo.Name == "<none>:<none>"
			isTaglessImage := !isDanglingImage && imgInfo.Tag == "<none>"

			if isDanglingImage || isTaglessImage {
				imageReferences = append(imageReferences, imgInfo.ID)
			} else {
				imageReferences = append(imageReferences, imgInfo.Name)
			}
		}
	}

	if err := imageReferencesRemove(imageReferences, rmiForce); err != nil {
		return err
	}

	return nil
}

func imageReferencesRemove(references []string, rmiForce bool) error {
	if len(references) == 0 {
		return nil
	}

	var args []string
	if rmiForce {
		args = append(args, "--force")
	}
	args = append(args, references...)

	if err := docker.CliRmi_LiveOutput(args...); err != nil {
		return err
	}

	return nil
}
