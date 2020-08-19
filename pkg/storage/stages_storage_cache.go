package storage

import (
	"context"

	"github.com/werf/werf/pkg/image"
)

type StagesStorageCache interface {
	GetAllStages(ctx context.Context, projectName string) (bool, []image.StageID, error)
	DeleteAllStages(ctx context.Context, projectName string) error
	GetStagesBySignature(ctx context.Context, projectName, signature string) (bool, []image.StageID, error)
	StoreStagesBySignature(ctx context.Context, projectName, signature string, stages []image.StageID) error
	DeleteStagesBySignature(ctx context.Context, projectName, signature string) error

	String() string
}
