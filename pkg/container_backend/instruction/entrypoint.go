package instruction

import (
	"context"
	"fmt"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_backend"
)

type Entrypoint struct {
	instructions.EntrypointCommand
	EntrypointResetCMD bool
}

func NewEntrypoint(i instructions.EntrypointCommand, entrypointResetCMD bool) *Entrypoint {
	return &Entrypoint{EntrypointCommand: i, EntrypointResetCMD: entrypointResetCMD}
}

func (i *Entrypoint) UsesBuildContext() bool {
	return false
}

func (i *Entrypoint) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContextArchive container_backend.BuildContextArchiver) error {
	if err := drv.Config(ctx, containerName, buildah.ConfigOpts{
		CommonOpts:             drvOpts,
		Entrypoint:             i.CmdLine,
		EntrypointPrependShell: i.PrependShell,
		EntrypointResetCMD:     i.EntrypointResetCMD,
	}); err != nil {
		return fmt.Errorf("error setting entrypoint %v for container %s: %w", i.CmdLine, containerName, err)
	}

	return nil
}
