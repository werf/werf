package instruction

import (
	"context"
	"fmt"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_backend"
)

type Cmd struct {
	instructions.CmdCommand
}

func NewCmd(i instructions.CmdCommand) *Cmd {
	return &Cmd{CmdCommand: i}
}

func (i *Cmd) UsesBuildContext() bool {
	return false
}

func (i *Cmd) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContextArchive container_backend.BuildContextArchiver) error {
	if err := drv.Config(ctx, containerName, buildah.ConfigOpts{
		CommonOpts:      drvOpts,
		Cmd:             i.CmdLine,
		CmdPrependShell: i.PrependShell,
	}); err != nil {
		return fmt.Errorf("error setting cmd %v for container %s: %w", i.CmdLine, containerName, err)
	}

	return nil
}
