package instruction

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_backend/build_context"
	dockerfile_instruction "github.com/werf/werf/pkg/dockerfile/instruction"
)

type Entrypoint struct {
	dockerfile_instruction.Entrypoint
}

func NewEntrypoint(i dockerfile_instruction.Entrypoint) *Entrypoint {
	return &Entrypoint{Entrypoint: i}
}

func (i *Entrypoint) UsesBuildContext() bool {
	return false
}

func (i *Entrypoint) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContext *build_context.BuildContext) error {
	if err := drv.Config(ctx, containerName, buildah.ConfigOpts{CommonOpts: drvOpts, Entrypoint: i.Entrypoint.Entrypoint}); err != nil {
		return fmt.Errorf("error setting entrypoint %v for container %s: %w", i.Entrypoint, containerName, err)
	}
	return nil
}
