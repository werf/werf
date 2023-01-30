package instruction

import (
	"context"
	"fmt"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_backend"
)

type StopSignal struct {
	instructions.StopSignalCommand
}

func NewStopSignal(i instructions.StopSignalCommand) *StopSignal {
	return &StopSignal{StopSignalCommand: i}
}

func (i *StopSignal) UsesBuildContext() bool {
	return false
}

func (i *StopSignal) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContextArchive container_backend.BuildContextArchiver) error {
	if err := drv.Config(ctx, containerName, buildah.ConfigOpts{CommonOpts: drvOpts, StopSignal: i.Signal}); err != nil {
		return fmt.Errorf("error setting stop signal %v for container %s: %w", i.Signal, containerName, err)
	}
	return nil
}
