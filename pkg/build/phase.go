package build

import "github.com/flant/werf/pkg/build/stage"

type Phase interface {
	Name() string
	BeforeImages() error
	AfterImages() error
	BeforeImageStages(img *Image) error
	OnImageStage(img *Image, stg stage.Interface) (bool, error)
	AfterImageStages(img *Image) error
	ImageProcessingShouldBeStopped(img *Image) bool
}

type BasePhase struct {
	Conveyor *Conveyor
}
