package instruction

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_backend/build_context"
)

type Volume struct {
	Volumes []string
}

func NewVolume(volumes []string) *Volume {
	return &Volume{Volumes: volumes}
}

func (i *Volume) UsesBuildContext() bool {
	return false
}

func (i *Volume) Name() string {
	return "VOLUME"
}

func (i *Volume) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContext *build_context.BuildContext) error {
	if err := drv.Config(ctx, containerName, buildah.ConfigOpts{CommonOpts: drvOpts, Volumes: i.Volumes}); err != nil {
		return fmt.Errorf("error setting volumes %v for container %s: %w", i.Volumes, containerName, err)
	}
	return nil
}
