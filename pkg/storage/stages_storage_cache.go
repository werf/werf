package storage

import "github.com/flant/werf/pkg/image"

type StagesStorageCache interface {
	GetAllStages(projectName string) (bool, []*image.Info, error)
	GetStagesBySignature(projectName, signature string) (bool, []*image.Info, error)
	StoreStagesBySignature(projectName, signature string, imageInfo []*image.Info) error
	DeleteStagesBySignature(projectName, signature string) error
}
