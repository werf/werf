package storage

import (
	"context"

	"github.com/werf/werf/v2/pkg/image"
)

type PrimaryStagesStorage interface {
	StagesStorage

	GetStageCustomTagMetadataIDs(ctx context.Context, opts ...Option) ([]string, error)
	GetStageCustomTagMetadata(ctx context.Context, tagOrID string) (*CustomTagMetadata, error)
	RegisterStageCustomTag(ctx context.Context, projectName string, stageDesc *image.StageDesc, tag string) error
	UnregisterStageCustomTag(ctx context.Context, tag string) error
}
