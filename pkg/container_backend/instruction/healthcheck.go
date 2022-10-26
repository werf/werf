package instruction

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/buildah/thirdparty"
	"github.com/werf/werf/pkg/container_backend"
	dockerfile_instruction "github.com/werf/werf/pkg/dockerfile/instruction"
)

type Healthcheck struct {
	dockerfile_instruction.Healthcheck
}

func NewHealthcheck(i dockerfile_instruction.Healthcheck) *Healthcheck {
	return &Healthcheck{Healthcheck: i}
}

func (i *Healthcheck) UsesBuildContext() bool {
	return false
}

func (i *Healthcheck) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContextArchive container_backend.BuildContextArchiver) error {
	if err := drv.Config(ctx, containerName, buildah.ConfigOpts{
		CommonOpts:  drvOpts,
		Healthcheck: (*thirdparty.HealthConfig)(i.Config),
	}); err != nil {
		return fmt.Errorf("error setting healthcheck for container %s: %w", containerName, err)
	}

	return nil
}
