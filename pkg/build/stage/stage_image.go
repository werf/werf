package stage

import (
	"github.com/werf/werf/pkg/container_backend"
	"github.com/werf/werf/pkg/container_backend/stage_builder"
)

type StageImage struct {
	Image   container_backend.LegacyImageInterface
	Builder stage_builder.StageBuilderInterface
}

func NewStageImage(containerBackend container_backend.ContainerBackend, baseImage string, image container_backend.LegacyImageInterface) *StageImage {
	return &StageImage{
		Image:   image,
		Builder: stage_builder.NewStageBuilder(containerBackend, baseImage, image),
	}
}
