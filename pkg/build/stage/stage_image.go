package stage

import (
	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/container_runtime/stage_builder"
)

type StageImage struct {
	Image   container_runtime.LegacyImageInterface
	Builder stage_builder.StageBuilderInterface
}

func NewStageImage(containerRuntime container_runtime.ContainerRuntime, fromImage container_runtime.ImageInterface, image container_runtime.LegacyImageInterface) *StageImage {
	return &StageImage{
		Image:   image,
		Builder: stage_builder.NewStageBuilder(containerRuntime, fromImage, image),
	}
}
