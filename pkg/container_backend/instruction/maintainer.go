package instruction

import (
	"context"
	"fmt"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_backend"
)

type Maintainer struct {
	instructions.MaintainerCommand
}

func NewMaintainer(i instructions.MaintainerCommand) *Maintainer {
	return &Maintainer{MaintainerCommand: i}
}

func (i *Maintainer) UsesBuildContext() bool {
	return false
}

func (i *Maintainer) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContextArchive container_backend.BuildContextArchiver) error {
	if err := drv.Config(ctx, containerName, buildah.ConfigOpts{
		CommonOpts: drvOpts,
		Maintainer: i.Maintainer,
	}); err != nil {
		return fmt.Errorf("error setting maintainer for container %s: %w", containerName, err)
	}

	return nil
}
