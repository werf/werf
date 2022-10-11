package instruction

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_backend/build_context"
	dockerfile_instruction "github.com/werf/werf/pkg/dockerfile/instruction"
)

type Expose struct {
	dockerfile_instruction.Expose
}

func NewExpose(i dockerfile_instruction.Expose) *Expose {
	return &Expose{Expose: i}
}

func (i *Expose) UsesBuildContext() bool {
	return false
}

func (i *Expose) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContext *build_context.BuildContext) error {
	if err := drv.Config(ctx, containerName, buildah.ConfigOpts{CommonOpts: drvOpts, Expose: i.Ports}); err != nil {
		return fmt.Errorf("error setting exposed ports %v for container %s: %w", i.Ports, containerName, err)
	}
	return nil
}
