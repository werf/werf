package stage

import (
	"github.com/werf/werf/pkg/build/builder"
	"github.com/werf/werf/pkg/container_runtime"
)

type StageImage struct {
	Image                container_runtime.LegacyImageInterface
	StageBuilderAccessor builder.StageBuilderAccessor
}

func NewStageImage(containerRuntime container_runtime.ContainerRuntime, image container_runtime.LegacyImageInterface) *StageImage {
	return &StageImage{
		Image:                image,
		StageBuilderAccessor: builder.NewStageBuilderAccessor(containerRuntime, image),
	}
}
