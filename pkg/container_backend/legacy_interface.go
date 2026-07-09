package container_backend

import (
	"context"

	"github.com/werf/werf/v2/pkg/image"
)

//go:generate mockgen -source legacy_interface.go -package mock -destination ../../test/mock/legacy_interface.go

type LegacyImageInterface interface {
	Name() string
	SetName(name string)

	GetTargetPlatform() string

	Pull(ctx context.Context) error
	Push(ctx context.Context) error

	SetBuildServiceLabels(labels map[string]string)
	GetBuildServiceLabels() map[string]string

	SetBuiltID(builtID string)
	BuiltID() string

	SetInfo(info *image.Info)

	IsExistsLocally() bool

	SetStageDesc(*image.StageDesc)
	GetStageDesc() *image.StageDesc

	GetFinalStageDesc() *image.StageDesc
	SetFinalStageDesc(*image.StageDesc)

	GetCopy() LegacyImageInterface
}
