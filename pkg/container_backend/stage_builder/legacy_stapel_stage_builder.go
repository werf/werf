package stage_builder

import (
	"context"
	"fmt"
	"strings"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/container_backend"
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

func (builder *LegacyStapelStageBuilder) Cleanup(ctx context.Context) error {
	if builder.Image.BuiltID() != "" {
		// FIXME: refactor or logic to prevent this dirty hack (prevention of final image deletion).
		if strings.HasPrefix(builder.Image.BuiltID(), "sha256:") {
			return nil
		}

		logboek.Context(ctx).Info().LogF("Cleanup built image %q\n", builder.Image.BuiltID())
		if err := builder.ContainerBackend.Rmi(ctx, builder.Image.BuiltID(), container_backend.RmiOpts{}); err != nil {
			return fmt.Errorf("unable to remove built dockerfile image %q: %w", builder.Image.BuiltID(), err)
		}
	}

	return nil
}
