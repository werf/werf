package instruction

import (
	"context"
	"fmt"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_backend"
)

type Expose struct {
	instructions.ExposeCommand
}

func NewExpose(i instructions.ExposeCommand) *Expose {
	return &Expose{ExposeCommand: i}
}

func (i *Expose) UsesBuildContext() bool {
	return false
}

func (i *Expose) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContextArchive container_backend.BuildContextArchiver) error {
	if err := drv.Config(ctx, containerName, buildah.ConfigOpts{CommonOpts: drvOpts, Expose: i.Ports}); err != nil {
		return fmt.Errorf("error setting exposed ports %v for container %s: %w", i.Ports, containerName, err)
	}
	return nil
}
