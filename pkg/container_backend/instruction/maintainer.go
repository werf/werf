package instruction

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_backend"
	dockerfile_instruction "github.com/werf/werf/pkg/dockerfile/instruction"
)

type Maintainer struct {
	dockerfile_instruction.Maintainer
}

func NewMaintainer(i dockerfile_instruction.Maintainer) *Maintainer {
	return &Maintainer{Maintainer: i}
}

func (i *Maintainer) UsesBuildContext() bool {
	return false
}

func (i *Maintainer) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContextArchive container_backend.BuildContextArchiver) error {
	if err := drv.Config(ctx, containerName, buildah.ConfigOpts{
		CommonOpts: drvOpts,
		Maintainer: i.Maintainer.Maintainer,
	}); err != nil {
		return fmt.Errorf("error setting maintainer for container %s: %w", containerName, err)
	}

	return nil
}
