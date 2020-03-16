package storage

import (
	"fmt"
	"strings"

	"github.com/flant/werf/pkg/container_runtime"
	"github.com/flant/werf/pkg/docker"

	"github.com/flant/logboek"
	"github.com/flant/werf/pkg/docker_registry"
)

const (
	RepoManagedImageRecord_ImageTagPrefix  = "managed-image-"
	RepoManagedImageRecord_ImageTagFormat  = "managed-image-%s"
	RepoManagedImageRecord_ImageNameFormat = "%s:managed-image-%s"

	RepoStage_ImageTagFormat = "stage-%s-%s"
)

type RepoStagesStorage struct {
	StagesStorage    // FIXME
	RepoAddress      string
	DockerRegistry   docker_registry.DockerRegistry
	ContainerRuntime container_runtime.ContainerRuntime
}

func NewRepoStagesStorage(repoAddress string, dockerRegistry docker_registry.DockerRegistry, containerRuntime container_runtime.ContainerRuntime) *RepoStagesStorage {
	return &RepoStagesStorage{
		RepoAddress:      repoAddress,
		DockerRegistry:   dockerRegistry,
		ContainerRuntime: containerRuntime,
	}
}

// TODO: Реализация интерфейса StagesStorage через низкоуровневый объект DockerRegistry

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

func (storage *RepoStagesStorage) StoreImage(image container_runtime.Image) error {
	switch containerRuntime := storage.ContainerRuntime.(type) {
	case *container_runtime.LocalDockerServerRuntime:
		return containerRuntime.ExportBuiltImage(image)
		// TODO: case *container_runtime.LocalHostRuntime:
	default:
		panic("not implemented")
	}
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
