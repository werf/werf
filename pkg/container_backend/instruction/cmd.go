package instruction

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_backend"
	dockerfile_instruction "github.com/werf/werf/pkg/dockerfile/instruction"
)

type Cmd struct {
	dockerfile_instruction.Cmd
}

func NewCmd(i dockerfile_instruction.Cmd) *Cmd {
	return &Cmd{Cmd: i}
}

func (i *Cmd) UsesBuildContext() bool {
	return false
}

func (i *Cmd) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContextArchive container_backend.BuildContextArchiver) error {
	if err := drv.Config(ctx, containerName, buildah.ConfigOpts{
		CommonOpts:      drvOpts,
		Cmd:             i.Cmd.Cmd,
		CmdPrependShell: i.PrependShell,
	}); err != nil {
		return fmt.Errorf("error setting cmd %v for container %s: %w", i.Cmd, containerName, err)
	}

	return nil
}
