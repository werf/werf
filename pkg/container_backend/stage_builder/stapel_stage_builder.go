package stage_builder

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/container_backend"
)

type StapelStageBuilderInterface interface {
	container_backend.BuildStapelStageOptionsInterface
	Build(ctx context.Context, opts container_backend.BuildOptions) error
}

type StapelStageBuilder struct {
	container_backend.BuildStapelStageOptions

	ContainerBackend container_backend.ContainerBackend
	FromImage        container_backend.ImageInterface
	Image            container_backend.ImageInterface
}

func NewStapelStageBuilder(containerBackend container_backend.ContainerBackend, fromImage, image container_backend.ImageInterface) *StapelStageBuilder {
	return &StapelStageBuilder{
		ContainerBackend: containerBackend,
		FromImage:        fromImage,
		Image:            image,
	}
}

func (builder *StapelStageBuilder) Build(ctx context.Context, opts container_backend.BuildOptions) error {
	// TODO: support introspect options

	builder.SetBaseImage(builder.FromImage.Name())

	builtID, err := builder.ContainerBackend.BuildStapelStage(ctx, builder.BuildStapelStageOptions)
	if err != nil {
		return fmt.Errorf("error building stapel stage with %s: %w", builder.ContainerBackend.String(), err)
	}

	builder.Image.SetBuiltID(builtID)

	return nil
}
