package build

import (
	"context"

	"github.com/werf/werf/pkg/build/stage"
)

type Phase interface {
	Name() string
	BeforeImages(ctx context.Context) error
	AfterImages(ctx context.Context) error
	BeforeImageStages(ctx context.Context, img *Image) error
	OnImageStage(ctx context.Context, img *Image, stg stage.Interface) error
	AfterImageStages(ctx context.Context, img *Image) error
	ImageProcessingShouldBeStopped(ctx context.Context, img *Image) bool
	Clone() Phase
}
