package stage_builder

import (
	"context"

	"github.com/werf/werf/pkg/container_backend"
)

type LegacyStapelStageBuilderInterface interface {
	Container() container_backend.LegacyContainer
	BuilderContainer() container_backend.LegacyBuilderContainer
	Build(ctx context.Context, opts container_backend.BuildOptions) error
}

type LegacyStapelStageBuilder struct {
	ContainerBackend container_backend.ContainerBackend
	Image            container_backend.LegacyImageInterface
}

func NewLegacyStapelStageBuilder(containerBackend container_backend.ContainerBackend, image container_backend.LegacyImageInterface) *LegacyStapelStageBuilder {
	return &LegacyStapelStageBuilder{
		ContainerBackend: containerBackend,
		Image:            image,
	}
}

func (builder *LegacyStapelStageBuilder) Container() container_backend.LegacyContainer {
	return builder.Image.Container()
}

func (builder *LegacyStapelStageBuilder) BuilderContainer() container_backend.LegacyBuilderContainer {
	return builder.Image.BuilderContainer()
}

func (builder *LegacyStapelStageBuilder) Build(ctx context.Context, opts container_backend.BuildOptions) error {
	return builder.Image.Build(ctx, opts)
}
