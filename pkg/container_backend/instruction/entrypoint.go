package instruction

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_backend"
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

func (i *Entrypoint) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContextArchive container_backend.BuildContextArchiver) error {
	if err := drv.Config(ctx, containerName, buildah.ConfigOpts{
		CommonOpts:             drvOpts,
		Entrypoint:             i.Entrypoint.Entrypoint,
		EntrypointPrependShell: i.PrependShell,
	}); err != nil {
		return fmt.Errorf("error setting entrypoint %v for container %s: %w", i.Entrypoint, containerName, err)
	}

	return nil
}
