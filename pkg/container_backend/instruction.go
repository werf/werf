package container_backend

import (
	"context"

	"github.com/werf/werf/v2/pkg/buildkit"
)

type InstructionInterface interface {
	Name() string
	ApplyBuildkit(ctx context.Context, stage *buildkit.DockerfileStageState) error
	UsesBuildContext() bool
}
