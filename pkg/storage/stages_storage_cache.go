package storage

import "github.com/werf/werf/pkg/image"

type StagesStorageCache interface {
	GetAllStages(projectName string) (bool, []image.StageID, error)
	DeleteAllStages(projectName string) error
	GetStagesBySignature(projectName, signature string) (bool, []image.StageID, error)
	StoreStagesBySignature(projectName, signature string, stages []image.StageID) error
	DeleteStagesBySignature(projectName, signature string) error

	String() string
}
