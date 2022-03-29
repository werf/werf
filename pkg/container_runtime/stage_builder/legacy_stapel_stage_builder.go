package stage_builder

import (
	"context"

	"github.com/werf/werf/pkg/container_runtime"
)

type LegacyStapelStageBuilderInterface interface {
	Container() container_runtime.LegacyContainer
	BuilderContainer() container_runtime.LegacyBuilderContainer
	Build(ctx context.Context, opts container_runtime.BuildOptions) error
}

type LegacyStapelStageBuilder struct {
	ContainerRuntime container_runtime.ContainerRuntime
	Image            container_runtime.LegacyImageInterface
}

func NewLegacyStapelStageBuilder(containerRuntime container_runtime.ContainerRuntime, image container_runtime.LegacyImageInterface) *LegacyStapelStageBuilder {
	return &LegacyStapelStageBuilder{
		ContainerRuntime: containerRuntime,
		Image:            image,
	}
}

func (builder *LegacyStapelStageBuilder) Container() container_runtime.LegacyContainer {
	return builder.Image.Container()
}

func (builder *LegacyStapelStageBuilder) BuilderContainer() container_runtime.LegacyBuilderContainer {
	return builder.Image.BuilderContainer()
}

func (builder *LegacyStapelStageBuilder) Build(ctx context.Context, opts container_runtime.BuildOptions) error {
	return builder.Image.Build(ctx, opts)
}
