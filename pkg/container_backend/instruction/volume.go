package instruction

import (
	"context"
	"fmt"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_backend"
)

type Volume struct {
	instructions.VolumeCommand
}

func NewVolume(i instructions.VolumeCommand) *Volume {
	return &Volume{VolumeCommand: i}
}

func (i *Volume) UsesBuildContext() bool {
	return false
}

func (i *Volume) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContextArchive container_backend.BuildContextArchiver) error {
	if err := drv.Config(ctx, containerName, buildah.ConfigOpts{CommonOpts: drvOpts, Volumes: i.Volumes}); err != nil {
		return fmt.Errorf("error setting volumes %v for container %s: %w", i.Volumes, containerName, err)
	}
	return nil
}
