package build

import "github.com/flant/werf/pkg/build/stage"

type Phase interface {
	Name() string
	OnStart(img *Image) error
	HandleStage(img *Image, stg stage.Interface) error
}
