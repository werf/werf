package instruction

import (
	"context"
	"fmt"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_backend"
)

type Env struct {
	instructions.EnvCommand
}

func NewEnv(i instructions.EnvCommand) *Env {
	return &Env{EnvCommand: i}
}

func (i *Env) UsesBuildContext() bool {
	return false
}

func (i *Env) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContextArchive container_backend.BuildContextArchiver) error {
	envs := make(map[string]string)
	for _, item := range i.Env {
		envs[item.Key] = item.Value
	}

	if err := drv.Config(ctx, containerName, buildah.ConfigOpts{CommonOpts: drvOpts, Envs: envs}); err != nil {
		return fmt.Errorf("error setting envs %v for container %s: %w", envs, containerName, err)
	}
	return nil
}
