package storage

import (
	"github.com/flant/werf/pkg/container_runtime"
	"github.com/flant/werf/pkg/image"
)

const (
	LocalStagesStorageAddress = ":local"
	NamelessImageRecordTag    = "__nameless__"
)

type StagesStorage interface {
	ConstructStageImageName(projectName, signature, uniqueID string) string

	GetRepoImages(projectName string) ([]*image.Info, error)
	DeleteRepoImage(options DeleteRepoImageOptions, repoImageList ...*image.Info) error

	GetRepoImagesBySignature(projectName, signature string) ([]*image.Info, error)

	GetImageInfo(stageImageName string) (*image.Info, error)

	// FetchImage will create a local image in the container-runtime
	FetchImage(img container_runtime.Image) error
	// StoreImage will store a local image into the container-runtime, local built image should exist prior running store
	StoreImage(img container_runtime.Image) error
	// CleanupLocalImage will remove a local image from container-runtime
	CleanupLocalImage(img container_runtime.Image) error
	ShouldFetchImage(img container_runtime.Image) (bool, error)
	ShouldCleanupLocalImage(img container_runtime.Image) (bool, error)

	AddManagedImage(projectName, imageName string) error
	RmManagedImage(projectName, imageName string) error
	GetManagedImages(projectName string) ([]string, error)

	Validate() error
	String() string
}

type DeleteRepoImageOptions struct {
	RmiForce                 bool
	SkipUsedImage            bool
	RmForce                  bool
	RmContainersThatUseImage bool
}

type StagesStorageOptions struct {
	RepoStagesStorageOptions
}

func NewStagesStorage(stagesStorageAddress string, containerRuntime container_runtime.ContainerRuntime, options StagesStorageOptions) (StagesStorage, error) {
	if stagesStorageAddress == LocalStagesStorageAddress {
		return NewLocalDockerServerStagesStorage(containerRuntime.(*container_runtime.LocalDockerServerRuntime)), nil
	} else { // Docker registry based stages storage
		return NewRepoStagesStorage(stagesStorageAddress, containerRuntime, options.RepoStagesStorageOptions)
	}
}
