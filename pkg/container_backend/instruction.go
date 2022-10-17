package container_backend

import (
	"context"

	"github.com/werf/werf/pkg/buildah"
)

type InstructionInterface interface {
	Name() string
	Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContextArchive BuildContextArchiver) error
	UsesBuildContext() bool
}
