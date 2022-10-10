package container_backend

import (
	"context"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_backend/build_context"
)

type InstructionInterface interface {
	Name() string
	Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContext *build_context.BuildContext) error
	UsesBuildContext() bool
}
