package instruction

import (
	"context"
	"fmt"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_backend"
)

type OnBuild struct {
	instructions.OnbuildCommand
}

func NewOnBuild(i instructions.OnbuildCommand) *OnBuild {
	return &OnBuild{OnbuildCommand: i}
}

func (i *OnBuild) UsesBuildContext() bool {
	return false
}

func (i *OnBuild) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContextArchive container_backend.BuildContextArchiver) error {
	if err := drv.Config(ctx, containerName, buildah.ConfigOpts{CommonOpts: drvOpts, OnBuild: i.Expression}); err != nil {
		return fmt.Errorf("error setting onbuild %v for container %s: %w", i.Expression, containerName, err)
	}
	return nil
}
