package instruction

import (
	"context"
	"fmt"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/buildah/thirdparty"
	"github.com/werf/werf/pkg/container_backend"
)

type Healthcheck struct {
	instructions.HealthCheckCommand
}

func NewHealthcheck(i instructions.HealthCheckCommand) *Healthcheck {
	return &Healthcheck{HealthCheckCommand: i}
}

func (i *Healthcheck) UsesBuildContext() bool {
	return false
}

func (i *Healthcheck) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContextArchive container_backend.BuildContextArchiver) error {
	healthCheck := &thirdparty.BuildahHealthConfig{
		Test:        i.Health.Test,
		Interval:    i.Health.Interval,
		Timeout:     i.Health.Timeout,
		StartPeriod: i.Health.StartPeriod,
		Retries:     i.Health.Retries,
	}

	if err := drv.Config(ctx, containerName, buildah.ConfigOpts{
		CommonOpts:  drvOpts,
		Healthcheck: healthCheck,
	}); err != nil {
		return fmt.Errorf("error setting healthcheck for container %s: %w", containerName, err)
	}

	return nil
}
