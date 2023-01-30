package instruction

import (
	"context"
	"fmt"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_backend"
)

type Workdir struct {
	instructions.WorkdirCommand
}

func NewWorkdir(i instructions.WorkdirCommand) *Workdir {
	return &Workdir{WorkdirCommand: i}
}

func (i *Workdir) UsesBuildContext() bool {
	return false
}

func (i *Workdir) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContextArchive container_backend.BuildContextArchiver) error {
	if err := drv.Config(ctx, containerName, buildah.ConfigOpts{CommonOpts: drvOpts, Workdir: i.Path}); err != nil {
		return fmt.Errorf("error setting workdir %s for container %s: %w", i.Path, containerName, err)
	}
	return nil
}
