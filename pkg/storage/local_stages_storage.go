package storage

import (
	"fmt"
	"strings"

	"github.com/golang/example/stringutil"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"

	"github.com/flant/logboek"

	"github.com/flant/werf/pkg/container_runtime"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/image"
)

const (
	LocalStage_ImageRepoPrefix = "werf-stages-storage/"
	LocalStage_ImageRepoFormat = "werf-stages-storage/%s"
	LocalStage_ImageFormat     = "werf-stages-storage/%s:%s-%s"

	LocalManagedImageRecord_ImageNamePrefix = "werf-managed-images/"
	LocalManagedImageRecord_ImageNameFormat = "werf-managed-images/%s"
	LocalManagedImageRecord_ImageFormat     = "werf-managed-images/%s:%s"
)

type LocalStagesStorage struct {
	// Local stages storage is compatible only with docker-server backed runtime
	LocalDockerServerRuntime *container_runtime.LocalDockerServerRuntime
}

func NewLocalStagesStorage(localDockerServerRuntime *container_runtime.LocalDockerServerRuntime) *LocalStagesStorage {
	return &LocalStagesStorage{LocalDockerServerRuntime: localDockerServerRuntime}
}

func (storage *LocalStagesStorage) Validate() error {
	return nil
}

func (storage *LocalStagesStorage) ConstructStageImageName(projectName, signature, uniqueID string) string {
	return fmt.Sprintf(LocalStage_ImageFormat, projectName, signature, uniqueID)
}

func (storage *LocalStagesStorage) GetRepoImages(projectName string) ([]*image.Info, error) {
	filterSet := localStagesStorageFilterSetBase(projectName)
	images, err := docker.Images(types.ImageListOptions{Filters: filterSet})
	if err != nil {
		return nil, fmt.Errorf("unable to get docker images: %s", err)
	}

	repoImageList := convertToRepoImageList(images)

	return repoImageList, nil
}

func (storage *LocalStagesStorage) DeleteRepoImage(options DeleteRepoImageOptions, repoImageList ...*image.Info) error {
	var err error
	repoImageList, err = processRelatedContainers(repoImageList, processRelatedContainersOptions{
		skipUsedImages:           options.SkipUsedImage,
		rmContainersThatUseImage: options.RmContainersThatUseImage,
		rmForce:                  options.RmForce,
	})
	if err != nil {
		return err
	}

	if err := deleteRepoImageListInLocalStagesStorage(repoImageList, options.RmiForce); err != nil {
		return err
	}

	return nil
}

func makeLocalManagedImageRecord(projectName, imageName string) string {
	tag := imageName
	if imageName == "" {
		tag = NamelessImageRecordTag
	}
	return fmt.Sprintf(LocalManagedImageRecord_ImageFormat, projectName, tag)
}

func (storage *LocalStagesStorage) GetImageInfo(stageImageName string) (*image.Info, error) {
	if inspect, err := storage.LocalDockerServerRuntime.GetImageInspect(stageImageName); err != nil {
		return nil, fmt.Errorf("unable to get image %s inspect: %s", stageImageName, err)
	} else if inspect != nil {
		return image.NewInfoFromInspect(stageImageName, inspect), nil
	} else {
		return nil, nil
	}
}

func (storage *LocalStagesStorage) AddManagedImage(projectName, imageName string) error {
	logboek.Debug.LogF("-- LocalStagesStorage.AddManagedImage %s %s\n", projectName, imageName)

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

func (storage *LocalStagesStorage) RmManagedImage(projectName, imageName string) error {
	logboek.Debug.LogF("-- LocalStagesStorage.RmManagedImage %s %s\n", projectName, imageName)

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

func (storage *LocalStagesStorage) GetManagedImages(projectName string) ([]string, error) {
	logboek.Debug.LogF("-- LocalStagesStorage.GetManagedImages %s\n", projectName)

	filterSet := filters.NewArgs()
	filterSet.Add("reference", fmt.Sprintf(LocalManagedImageRecord_ImageNameFormat, projectName))

	images, err := docker.Images(types.ImageListOptions{Filters: filterSet})
	if err != nil {
		return nil, fmt.Errorf("unable to get docker images: %s", err)
	}

	res := []string{}
	for _, img := range images {
		for _, repoTag := range img.RepoTags {
			tag := stringutil.Reverse(strings.SplitN(stringutil.Reverse(repoTag), ":", 2)[1])

			if tag == NamelessImageRecordTag {
				res = append(res, "")
			} else {
				res = append(res, tag)
			}
		}
	}
	return res, nil
}

func (storage *LocalStagesStorage) GetRepoImagesBySignature(projectName, signature string) ([]*image.Info, error) {
	filterSet := filters.NewArgs()
	filterSet.Add("reference", fmt.Sprintf(LocalStage_ImageRepoFormat, projectName))
	// NOTE signature already depends on build-cache-version
	filterSet.Add("label", fmt.Sprintf("%s=%s", image.WerfStageSignatureLabel, signature))

	images, err := docker.Images(types.ImageListOptions{Filters: filterSet})
	if err != nil {
		return nil, fmt.Errorf("unable to get docker images: %s", err)
	}

	repoImages := convertToRepoImageList(images)
	return repoImages, nil
}

func (storage *LocalStagesStorage) FetchImage(image container_runtime.Image) error {
	return nil
}

func (storage *LocalStagesStorage) StoreImage(image container_runtime.Image) error {
	return storage.LocalDockerServerRuntime.TagBuiltImageByName(image)
}

func (storage *LocalStagesStorage) CleanupLocalImage(image container_runtime.Image) error {
	return nil
}

func (storage *LocalStagesStorage) String() string {
	return ":local"
}

type processRelatedContainersOptions struct {
	skipUsedImages           bool
	rmContainersThatUseImage bool
	rmForce                  bool
}

func processRelatedContainers(repoImages []*image.Info, options processRelatedContainersOptions) ([]*image.Info, error) {
	filterSet := filters.NewArgs()
	for _, repoImage := range repoImages {
		filterSet.Add("ancestor", repoImage.ID)
	}

	containerList, err := containerListByFilterSet(filterSet)
	if err != nil {
		return nil, err
	}

	var repoImageListToExcept []*image.Info
	var containerListToRemove []types.Container
	for _, container := range containerList {
		for _, repoImage := range repoImages {
			if repoImage.ID == container.ImageID {
				if options.skipUsedImages {
					logboek.Default.LogFDetails("Skip image %s (used by container %s)\n", logImageName(repoImage), logContainerName(container))
					repoImageListToExcept = append(repoImageListToExcept, repoImage)
				} else if options.rmContainersThatUseImage {
					containerListToRemove = append(containerListToRemove, container)
				} else {
					return nil, fmt.Errorf("cannot remove image %s used by container %s\n%s", logImageName(repoImage), logContainerName(container), "Use --force option to remove all containers that are based on deleting werf docker images")
				}
			}
		}
	}

	if err := deleteContainers(containerListToRemove, options.rmForce); err != nil {
		return nil, err
	}

	repoImages = exceptRepoImageList(repoImages, repoImageListToExcept...)

	return repoImages, nil
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

func exceptRepoImageList(repoImageList []*image.Info, repoImageListToExcept ...*image.Info) []*image.Info {
	var newRepoImageList []*image.Info

loop:
	for _, repoImage := range repoImageList {
		for _, repoImageToExcept := range repoImageListToExcept {
			if repoImageToExcept == repoImage {
				continue loop
			}
		}

		newRepoImageList = append(newRepoImageList, repoImage)
	}

	return newRepoImageList
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

func convertToRepoImageList(imageSummaryList []types.ImageSummary) (repoImageList []*image.Info) {
	for _, imageSummary := range imageSummaryList {
		repoTags := imageSummary.RepoTags
		if len(repoTags) == 0 {
			repoTags = append(repoTags, "<none>:<none>")
		}

		for _, repoTag := range repoTags {
			parts := strings.SplitN(stringutil.Reverse(repoTag), ":", 2)
			repository := stringutil.Reverse(parts[0])
			tag := stringutil.Reverse(parts[1])

			repoImage := &image.Info{
				Repository: repository,
				Tag:        tag,
				ID:         imageSummary.ID,
				ParentID:   imageSummary.ParentID,
				Name:       repoTag,
				Labels:     imageSummary.Labels,
				Size:       imageSummary.Size,
			}

			repoImage.SetCreatedAtUnix(imageSummary.Created)

			repoImageList = append(repoImageList, repoImage)
		}
	}

	return repoImageList
}

func deleteRepoImageListInLocalStagesStorage(repoImageList []*image.Info, rmiForce bool) error {
	var imageReferences []string
	for _, repoImage := range repoImageList {
		if repoImage.Name == "" {
			imageReferences = append(imageReferences, repoImage.ID)
		} else {
			isDanglingImage := repoImage.Name == "<none>:<none>"
			isTaglessImage := !isDanglingImage && repoImage.Tag == "<none>"

			if isDanglingImage || isTaglessImage {
				imageReferences = append(imageReferences, repoImage.ID)
			} else {
				imageReferences = append(imageReferences, repoImage.Name)
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
