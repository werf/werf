package instruction

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_backend"
	dockerfile_instruction "github.com/werf/werf/pkg/dockerfile/instruction"
)

type Volume struct {
	dockerfile_instruction.Volume
}

func NewVolume(i dockerfile_instruction.Volume) *Volume {
	return &Volume{Volume: i}
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
