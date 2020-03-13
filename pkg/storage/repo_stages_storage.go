package storage

import (
	"fmt"

	"github.com/flant/werf/pkg/container_runtime"

	"github.com/flant/logboek"
	"github.com/flant/werf/pkg/docker_registry"
)

const (
	ManagedImageRecord_ImageNamePrefix = "werf-managed-images/"
	ManagedImageRecord_ImageNameFormat = "werf-managed-images/%s"
	ManagedImageRecord_ImageFormat     = "werf-managed-images/%s:%s"

	RepoManagedImageRecord_ImageTagFormat = "managed-image-%s"
	RepoStage_ImageTagFormat              = "stage-%s-%s"
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
	panic("no")
	logboek.Debug.LogF("-- RepoStagesStorage.AddManagedImage %s %s\n", projectName, imageName)

	fullImageName := fmt.Sprintf(RepoManagedImageRecord_ImageTagFormat, imageName)
	_ = fullImageName

	// Create manifest using docker registry

	//if exsts, err := docker.ImageExist(fullImageName); err != nil {
	//	return fmt.Errorf("unable to check existence of image %q: %s", fullImageName, err)
	//} else if exsts {
	//	return nil
	//}
	//
	//if err := docker.CreateImage(fullImageName); err != nil {
	//	return fmt.Errorf("unable to create image %q: %s", fullImageName, err)
	//}

	return nil
}

func (storage *RepoStagesStorage) RmManagedImage(projectName, imageName string) error {
	panic("no")
	return nil
}

func (storage *RepoStagesStorage) GetManagedImages(projectName string) ([]string, error) {
	panic("no")
	return nil, nil
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
