package synchronization_server

import (
	"context"

	"github.com/werf/werf/pkg/image"
)

type StagesStorageCacheInterface interface {
	GetAllStages(ctx context.Context, projectName string) (bool, []image.StageID, error)
	DeleteAllStages(ctx context.Context, projectName string) error
	GetStagesByDigest(ctx context.Context, projectName, digest string) (bool, []image.StageID, error)
	StoreStagesByDigest(ctx context.Context, projectName, digest string, stages []image.StageID) error
	DeleteStagesByDigest(ctx context.Context, projectName, digest string) error

	String() string
}
