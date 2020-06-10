package storage

import (
	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/image"
)

const (
	LocalStorageAddress             = ":local"
	DefaultKubernetesStorageAddress = "kubernetes://werf-synchronization"
	NamelessImageRecordTag          = "__nameless__"
)

type StagesStorage interface {
	GetAllStages(projectName string) ([]image.StageID, error)
	GetStagesBySignature(projectName, signature string) ([]image.StageID, error)
	GetStageDescription(projectName, signature string, uniqueID int64) (*image.StageDescription, error)
	DeleteStages(options DeleteImageOptions, stages ...*image.StageDescription) error

	ConstructStageImageName(projectName, signature string, uniqueID int64) string

	// FetchImage will create a local image in the container-runtime
	FetchImage(img container_runtime.Image) error
	// StoreImage will store a local image into the container-runtime, local built image should exist prior running store
	StoreImage(img container_runtime.Image) error
	ShouldFetchImage(img container_runtime.Image) (bool, error)

	CreateRepo() error
	DeleteRepo() error

	AddManagedImage(projectName, imageName string) error
	RmManagedImage(projectName, imageName string) error
	GetManagedImages(projectName string) ([]string, error)

	String() string
	Address() string
}

type StagesStorageOptions struct {
	RepoStagesStorageOptions
}

func NewStagesStorage(stagesStorageAddress string, containerRuntime container_runtime.ContainerRuntime, options StagesStorageOptions) (StagesStorage, error) {
	if stagesStorageAddress == LocalStorageAddress {
		return NewLocalDockerServerStagesStorage(containerRuntime.(*container_runtime.LocalDockerServerRuntime)), nil
	} else { // Docker registry based stages storage
		return NewRepoStagesStorage(stagesStorageAddress, containerRuntime, options.RepoStagesStorageOptions)
	}
}
