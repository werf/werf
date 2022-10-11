package instruction

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_backend/build_context"
	dockerfile_instruction "github.com/werf/werf/pkg/dockerfile/instruction"
)

type Shell struct {
	dockerfile_instruction.Shell
}

func NewShell(i dockerfile_instruction.Shell) *Shell {
	return &Shell{Shell: i}
}

func (i *Shell) UsesBuildContext() bool {
	return false
}

func (i *Shell) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContext *build_context.BuildContext) error {
	if err := drv.Config(ctx, containerName, buildah.ConfigOpts{CommonOpts: drvOpts, Shell: i.Shell.Shell}); err != nil {
		return fmt.Errorf("error setting shell %v for container %s: %w", i.Shell, containerName, err)
	}
	return nil
}
