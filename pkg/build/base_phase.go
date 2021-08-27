package build

import (
	"context"

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

func (phase *BasePhase) BeforeImageStages(_ context.Context, _ *Image) error {
	return nil
}

func (phase *BasePhase) OnImageStage(_ context.Context, _ *Image, _ stage.Interface) error {
	return nil
}

func (phase *BasePhase) AfterImageStages(ctx context.Context, img *Image) error {
	return nil
}

func (phase *BasePhase) ImageProcessingShouldBeStopped(_ context.Context, _ *Image) bool {
	return false
}
