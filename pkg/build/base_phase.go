package build

import (
	"context"

	"github.com/werf/werf/pkg/build/image"
	"github.com/werf/werf/pkg/build/stage"
)

type BasePhase struct {
	Conveyor *Conveyor
}

func (phase *BasePhase) BeforeImages(_ context.Context) error {
	return nil
}

func (phase *BasePhase) AfterImages(_ context.Context) error {
	return nil
}

func (phase *BasePhase) BeforeImageStages(ctx context.Context, img *image.Image) (deferFn func(), err error) {
	return nil, nil
}

func (phase *BasePhase) OnImageStage(_ context.Context, _ *image.Image, _ stage.Interface) error {
	return nil
}

func (phase *BasePhase) AfterImageStages(ctx context.Context, img *image.Image) error {
	return nil
}

func (phase *BasePhase) ImageProcessingShouldBeStopped(_ context.Context, _ *image.Image) bool {
	return false
}
