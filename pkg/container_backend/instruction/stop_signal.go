package instruction

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_backend/build_context"
)

type StopSignal struct {
	Signal string
}

func NewStopSignal(signal string) *StopSignal {
	return &StopSignal{Signal: signal}
}

func (i *StopSignal) UsesBuildContext() bool {
	return false
}

func (i *StopSignal) Name() string {
	return "STOPSIGNAL"
}

func (i *StopSignal) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContext *build_context.BuildContext) error {
	if err := drv.Config(ctx, containerName, buildah.ConfigOpts{CommonOpts: drvOpts, StopSignal: i.Signal}); err != nil {
		return fmt.Errorf("error setting stop signal %v for container %s: %w", i.Signal, containerName, err)
	}
	return nil
}
