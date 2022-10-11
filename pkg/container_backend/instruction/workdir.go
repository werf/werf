package instruction

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_backend/build_context"
	dockerfile_instruction "github.com/werf/werf/pkg/dockerfile/instruction"
)

type Workdir struct {
	dockerfile_instruction.Workdir
}

func NewWorkdir(i dockerfile_instruction.Workdir) *Workdir {
	return &Workdir{Workdir: i}
}

func (i *Workdir) UsesBuildContext() bool {
	return false
}

func (i *Workdir) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContext *build_context.BuildContext) error {
	if err := drv.Config(ctx, containerName, buildah.ConfigOpts{CommonOpts: drvOpts, Workdir: i.Workdir.Workdir}); err != nil {
		return fmt.Errorf("error setting workdir %s for container %s: %w", i.Workdir, containerName, err)
	}
	return nil
}
