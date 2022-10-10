package instruction

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_backend/build_context"
)

type Shell struct {
	Shell []string
}

func NewShell(shell []string) *Shell {
	return &Shell{Shell: shell}
}

func (i *Shell) UsesBuildContext() bool {
	return false
}

func (i *Shell) Name() string {
	return "SHELL"
}

func (i *Shell) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContext *build_context.BuildContext) error {
	if err := drv.Config(ctx, containerName, buildah.ConfigOpts{CommonOpts: drvOpts, Shell: i.Shell}); err != nil {
		return fmt.Errorf("error setting shell %v for container %s: %w", i.Shell, containerName, err)
	}
	return nil
}
