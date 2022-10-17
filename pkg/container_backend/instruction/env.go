package instruction

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_backend"
	dockerfile_instruction "github.com/werf/werf/pkg/dockerfile/instruction"
)

type Env struct {
	dockerfile_instruction.Env
}

func NewEnv(i dockerfile_instruction.Env) *Env {
	return &Env{Env: i}
}

func (i *Env) UsesBuildContext() bool {
	return false
}

func (i *Env) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContextArchive container_backend.BuildContextArchiver) error {
	if err := drv.Config(ctx, containerName, buildah.ConfigOpts{CommonOpts: drvOpts, Envs: i.Envs}); err != nil {
		return fmt.Errorf("error setting envs %v for container %s: %w", i.Envs, containerName, err)
	}
	return nil
}
