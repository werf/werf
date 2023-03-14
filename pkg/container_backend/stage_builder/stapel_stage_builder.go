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
	BaseImage        string
	Image            container_backend.ImageInterface
}

func NewStapelStageBuilder(containerBackend container_backend.ContainerBackend, baseImage string, image container_backend.ImageInterface) *StapelStageBuilder {
	return &StapelStageBuilder{
		ContainerBackend: containerBackend,
		BaseImage:        baseImage,
		Image:            image,
	}
}

func (builder *StapelStageBuilder) Build(ctx context.Context, opts container_backend.BuildOptions) error {
	finalOpts := builder.BuildStapelStageOptions
	finalOpts.TargetPlatform = opts.TargetPlatform
	// TODO: support introspect options

	builtID, err := builder.ContainerBackend.BuildStapelStage(ctx, builder.BaseImage, finalOpts)
	if err != nil {
		return fmt.Errorf("error building stapel stage with %s: %w", builder.ContainerBackend.String(), err)
	}

	builder.Image.SetBuiltID(builtID)

	return nil
}
