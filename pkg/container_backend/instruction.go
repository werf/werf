package container_backend

import (
	"context"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"

	"github.com/werf/werf/v2/pkg/buildah"
)

type InstructionInterface interface {
	Name() string
	Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContextArchive BuildContextArchiver) error
	UsesBuildContext() bool
}

type MountsInterface interface {
	GetMounts() []*instructions.Mount
}
