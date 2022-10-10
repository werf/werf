package instruction

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_backend/build_context"
)

type Cmd struct {
	Cmd []string
}

func NewCmd(cmd []string) *Cmd {
	return &Cmd{Cmd: cmd}
}

func (i *Cmd) UsesBuildContext() bool {
	return false
}

func (i *Cmd) Name() string {
	return "CMD"
}

func (i *Cmd) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContext *build_context.BuildContext) error {
	if err := drv.Config(ctx, containerName, buildah.ConfigOpts{CommonOpts: drvOpts, Cmd: i.Cmd}); err != nil {
		return fmt.Errorf("error setting cmd %v for container %s: %w", i.Cmd, containerName, err)
	}
	return nil
}
