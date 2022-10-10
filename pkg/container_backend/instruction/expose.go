package instruction

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_backend/build_context"
)

type Expose struct {
	Ports []string
}

func NewExpose(ports []string) *Expose {
	return &Expose{Ports: ports}
}

func (i *Expose) UsesBuildContext() bool {
	return false
}

func (i *Expose) Name() string {
	return "EXPOSE"
}

func (i *Expose) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContext *build_context.BuildContext) error {
	if err := drv.Config(ctx, containerName, buildah.ConfigOpts{
		CommonOpts: drvOpts,
		Expose:     i.Ports,
	}); err != nil {
		return fmt.Errorf("error setting exposed ports %v for container %s: %w", i.Ports, containerName, err)
	}
	return nil
}
