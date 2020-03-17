package storage

import (
	"fmt"
	"strings"

	"github.com/flant/werf/pkg/image"

	"github.com/flant/werf/pkg/container_runtime"
	"github.com/flant/werf/pkg/docker"

	"github.com/flant/logboek"
	"github.com/flant/werf/pkg/docker_registry"
)

const (
	RepoStage_ImageTagFormat = "%s-%s"
	RepoStage_ImageFormat    = "%s:%s-%s"

	RepoManagedImageRecord_ImageTagPrefix  = "managed-image-"
	RepoManagedImageRecord_ImageTagFormat  = "managed-image-%s"
	RepoManagedImageRecord_ImageNameFormat = "%s:managed-image-%s"
)

type RepoStagesStorage struct {
	RepoAddress      string
	DockerRegistry   docker_registry.DockerRegistry
	ContainerRuntime container_runtime.ContainerRuntime
}

type RepoStagesStorageOptions struct {
	docker_registry.DockerRegistryOptions
	Implementation string
}

func NewRepoStagesStorage(repoAddress string, containerRuntime container_runtime.ContainerRuntime, options RepoStagesStorageOptions) (*RepoStagesStorage, error) {
	implementation := options.Implementation

	dockerRegistry, err := docker_registry.NewDockerRegistry(repoAddress, implementation, options.DockerRegistryOptions)
	if err != nil {
		return nil, fmt.Errorf("error creating docker registry accessor for repo %q: %s", repoAddress, err)
	}

	return &RepoStagesStorage{
		RepoAddress:      repoAddress,
		DockerRegistry:   dockerRegistry,
		ContainerRuntime: containerRuntime,
	}, nil
}

func (storage *RepoStagesStorage) ConstructStageImageName(projectName, signature, uniqueID string) string {
	return fmt.Sprintf(RepoStage_ImageFormat, storage.RepoAddress, signature, uniqueID)
}

func (storage *RepoStagesStorage) GetRepoImages(projectName string) ([]*image.Info, error) {
	return nil, fmt.Errorf("storage.RepoStagesStorage->GetRepoImages not implemented")
}

func (storage *RepoStagesStorage) DeleteRepoImage(options DeleteRepoImageOptions, repoImageList ...*image.Info) error {
	return fmt.Errorf("storage.RepoStagesStorage->DeleteRepoImage not implemented")
}

func (storage *RepoStagesStorage) GetRepoImagesBySignature(projectName, signature string) ([]*image.Info, error) {
	var res []*image.Info

	if tags, err := storage.DockerRegistry.Tags(storage.RepoAddress); err != nil {
		return nil, fmt.Errorf("unable to fetch tags for repo %q: %s", storage.RepoAddress, err)
	} else {
		logboek.Debug.LogF("-- RepoStagesStorage.GetRepoImagesBySignature fetched tags for %q: %#v\n", storage.RepoAddress, tags)
		for _, tag := range tags {
			if !strings.HasPrefix(tag, signature) {
				logboek.Debug.LogF("Discard tag %q: should have prefix %q\n", tag, signature)
				continue
			}

			logboek.Debug.LogF("Tag %q is suitable for signature %q\n", tag, signature)

			fullImageName := fmt.Sprintf("%s:%s", storage.RepoAddress, tag)
			if imgInfo, err := storage.DockerRegistry.GetRepoImage(fullImageName); err != nil {
				return nil, fmt.Errorf("unable to get image %q info from repo: %s", fullImageName, err)
			} else {
				logboek.Debug.LogF("Got imgInfo for %q: %#v\n", fullImageName, imgInfo)
				res = append(res, imgInfo)
			}
		}
	}

	logboek.Debug.LogF("-- RepoStagesStorage.GetRepoImagesBySignature result for %q: %#v\n", storage.RepoAddress, res)

	return res, nil
}

func (storage *RepoStagesStorage) GetImageInfo(stageImageName string) (*image.Info, error) {
	if imgInfo, err := storage.DockerRegistry.GetRepoImage(stageImageName); err != nil {
		return nil, fmt.Errorf("unable to get image %q info from repo: %s", stageImageName, err)
	} else {
		return imgInfo, nil
	}
}

func (storage *RepoStagesStorage) AddManagedImage(projectName, imageName string) error {
	logboek.Debug.LogF("-- RepoStagesStorage.AddManagedImage %s %s\n", projectName, imageName)

	fullImageName := makeRepoManagedImageRecord(storage.RepoAddress, imageName)

	if isExists, err := storage.DockerRegistry.IsRepoImageExists(fullImageName); err != nil {
		return err
	} else if isExists {
		logboek.Debug.LogF("-- RepoStagesStorage.AddManagedImage record %q is exists => exiting\n", fullImageName)
		return nil
	}

	logboek.Debug.LogF("-- RepoStagesStorage.AddManagedImage record %q does not exist => creating record\n", fullImageName)

	switch storage.ContainerRuntime.(type) {
	case *container_runtime.LocalDockerServerRuntime:
		if err := docker.CreateImage(fullImageName); err != nil {
			return fmt.Errorf("unable to create image %q: %s", fullImageName, err)
		}
		defer func() {
			if err := docker.CliRmi(fullImageName); err != nil {
				// TODO: errored repo state
				logboek.Error.LogF("unable to remove temporary image %q: %s", fullImageName, err)
			}
		}()

		if err := docker.CliPushWithRetries(fullImageName); err != nil {
			return fmt.Errorf("unable to push image %q: %s", fullImageName, err)
		}

		return nil
	default: // TODO: case *container_runtime.LocalHostRuntime:
		panic("not implemented")
	}
}

func (storage *RepoStagesStorage) RmManagedImage(projectName, imageName string) error {
	logboek.Debug.LogF("-- RepoStagesStorage.RmManagedImage %s %s\n", projectName, imageName)

	fullImageName := makeRepoManagedImageRecord(storage.RepoAddress, imageName)

	if imgInfo, err := storage.DockerRegistry.TryGetRepoImage(fullImageName); err != nil {
		return fmt.Errorf("unable to get repo image %q info: %s", fullImageName, err)
	} else if imgInfo == nil {
		logboek.Debug.LogF("-- RepoStagesStorage.RmManagedImage record %q does not exist => exiting\n", fullImageName)
		return nil
	} else {
		if err := storage.DockerRegistry.DeleteRepoImage(imgInfo); err != nil {
			return fmt.Errorf("unable to delete image %q from repo: %s", fullImageName, err)
		}
	}

	return nil
}

func (storage *RepoStagesStorage) GetManagedImages(projectName string) ([]string, error) {
	logboek.Debug.LogF("-- RepoStagesStorage.GetManagedImages %s %s\n", projectName)

	res := []string{}

	if tags, err := storage.DockerRegistry.Tags(storage.RepoAddress); err != nil {
		return nil, fmt.Errorf("unable to get repo %q tags: %s", storage.RepoAddress, err)
	} else {
		for _, tag := range tags {
			if !strings.HasPrefix(tag, RepoManagedImageRecord_ImageTagPrefix) {
				continue
			}

			managedImageName := strings.TrimPrefix(tag, RepoManagedImageRecord_ImageTagPrefix)
			if managedImageName == NamelessImageRecordTag {
				res = append(res, "")
			} else {
				res = append(res, managedImageName)
			}
		}
	}

	return res, nil
}

func (storage *RepoStagesStorage) FetchImage(image container_runtime.Image) error {
	switch containerRuntime := storage.ContainerRuntime.(type) {
	case *container_runtime.LocalDockerServerRuntime:
		// FIXME: construct image name
		// TODO: lock by name to prevent double pull of the same stage
		return containerRuntime.PullImageFromRegistry(image)
		// TODO: case *container_runtime.LocalHostRuntime:
	default:
		panic("not implemented")
	}
}

func (storage *RepoStagesStorage) StoreImage(image container_runtime.Image) error {
	switch containerRuntime := storage.ContainerRuntime.(type) {
	case *container_runtime.LocalDockerServerRuntime:
		// FIXME: construct image name
		return containerRuntime.PushBuiltImage(image)
		// TODO: case *container_runtime.LocalHostRuntime:
	default:
		panic("not implemented")
	}
}

func (storage *RepoStagesStorage) CleanupLocalImage(image container_runtime.Image) error {
	// FIXME
	logboek.Default.LogF("Cleaning local image %#v\n", image)
	return nil
}

func (storage *RepoStagesStorage) String() string {
	return fmt.Sprintf("repo stages storage (%q)", storage.RepoAddress)
}

func makeRepoManagedImageRecord(repoAddress, imageName string) string {
	tagSuffix := imageName
	if imageName == "" {
		tagSuffix = NamelessImageRecordTag
	}
	return fmt.Sprintf(RepoManagedImageRecord_ImageNameFormat, repoAddress, tagSuffix)
}
