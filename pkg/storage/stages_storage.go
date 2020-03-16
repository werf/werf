package storage

import (
	"fmt"

	"github.com/flant/werf/pkg/container_runtime"

	"github.com/flant/werf/pkg/docker_registry"
	"github.com/flant/werf/pkg/image"
)

const (
	LocalStagesStorageAddress = ":local"
	NamelessImageRecordTag    = "__nameless__"
)

type StagesStorage interface {
	GetRepoImages(projectName string) ([]*image.Info, error)
	DeleteRepoImage(options DeleteRepoImageOptions, repoImageList ...*image.Info) error

	GetRepoImagesBySignature(projectName, signature string) ([]*image.Info, error)

	// в том числе docker pull из registry + image.SyncDockerState
	// lock по имени image чтобы не делать 2 раза pull одновременно
	//SyncStageImage(stageImage image.ImageInterface) error
	//StoreStageImage(stageImage image.ImageInterface) error

	//FetchImage TODO
	StoreImage(image container_runtime.Image) error

	AddManagedImage(projectName, imageName string) error
	RmManagedImage(projectName, imageName string) error
	GetManagedImages(projectName string) ([]string, error)

	String() string
}

type DeleteRepoImageOptions struct {
	RmiForce                 bool
	SkipUsedImage            bool
	RmForce                  bool
	RmContainersThatUseImage bool
}

func NewStagesStorage(stagesStorageAddress string, containerRuntime container_runtime.ContainerRuntime) (StagesStorage, error) {
	if stagesStorageAddress == LocalStagesStorageAddress {
		return NewLocalStagesStorage(containerRuntime.(*container_runtime.LocalDockerServerRuntime)), nil
	} else { // Docker registry based stages storage
		if dockerRegistry, err := docker_registry.NewDockerRegistry(stagesStorageAddress); err != nil {
			return nil, fmt.Errorf("error creating docker registry accessor for repo %q: %s", stagesStorageAddress, err)
		} else {
			return NewRepoStagesStorage(stagesStorageAddress, dockerRegistry, containerRuntime), nil
		}
	}
}
