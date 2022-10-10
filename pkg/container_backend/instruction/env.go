package instruction

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_backend/build_context"
)

type Env struct {
	Envs map[string]string
}

func NewEnv(envs map[string]string) *Env {
	return &Env{Envs: envs}
}

func (i *Env) UsesBuildContext() bool {
	return false
}

func (i *Env) Name() string {
	return "ENV"
}

func (i *Env) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContext *build_context.BuildContext) error {
	if err := drv.Config(ctx, containerName, buildah.ConfigOpts{CommonOpts: drvOpts, Envs: i.Envs}); err != nil {
		return fmt.Errorf("error setting envs %v for container %s: %w", i.Envs, containerName, err)
	}
	return nil
}
