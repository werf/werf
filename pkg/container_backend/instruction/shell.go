package instruction

import (
	"context"
	"fmt"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_backend"
)

type Shell struct {
	instructions.ShellCommand
}

func NewShell(i instructions.ShellCommand) *Shell {
	return &Shell{ShellCommand: i}
}

func (i *Shell) UsesBuildContext() bool {
	return false
}

func (i *Shell) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContextArchive container_backend.BuildContextArchiver) error {
	if err := drv.Config(ctx, containerName, buildah.ConfigOpts{CommonOpts: drvOpts, Shell: i.Shell}); err != nil {
		return fmt.Errorf("error setting shell %v for container %s: %w", i.Shell, containerName, err)
	}
	return nil
}
