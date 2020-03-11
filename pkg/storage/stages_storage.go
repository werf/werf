package storage

import (
	"github.com/flant/werf/pkg/image"
)

type StagesStorage interface {
	GetRepoImages(projectName string) ([]*image.Info, error)
	DeleteRepoImage(options DeleteImageOptions, imageInfo ...*image.Info) error

	GetRepoImagesBySignature(projectName, signature string) ([]*image.Info, error)

	// в том числе docker pull из registry + image.SyncDockerState
	// lock по имени image чтобы не делать 2 раза pull одновременно
	SyncStageImage(stageImage image.ImageInterface) error
	StoreStageImage(stageImage image.ImageInterface) error

	AddManagedImage(projectName, imageName string) error
	RmManagedImage(projectName, imageName string) error
	GetManagedImages(projectName string) ([]string, error)

	String() string
}

type DeleteImageOptions struct {
	SkipUsedImages bool
	RmiForce       bool
	RmForce        bool
}
