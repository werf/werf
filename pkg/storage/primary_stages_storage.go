package storage

import (
	"context"

	"github.com/werf/werf/v3/pkg/image"
)

type PrimaryStagesStorage interface {
	StagesStorage

	GetStageCustomTagMetadataIDs(ctx context.Context, opts ...Option) ([]string, error)
	GetStageCustomTagMetadata(ctx context.Context, id string) (*CustomTagMetadata, error)
	RegisterStageCustomTag(ctx context.Context, projectName string, stageDesc *image.StageDesc, tag CustomTag) error
	UnregisterStageCustomTag(ctx context.Context, tag CustomTag) error
}
