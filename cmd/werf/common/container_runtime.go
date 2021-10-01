package common

import (
	"context"
	"fmt"
	"os"

	"github.com/werf/werf/pkg/docker"

	"github.com/werf/werf/pkg/buildah"

	"github.com/werf/werf/pkg/container_runtime"
)

func InitProcessContainerRuntime(ctx context.Context, cmdData *CmdData) (bool, container_runtime.ContainerRuntime, context.Context, error) {
	if v := os.Getenv("WERF_CONTAINER_RUNTIME_BUILDAH"); v != "" {
		mode := buildah.Mode(v)

		shouldTerminate, mode, err := buildah.InitProcess(func() (buildah.Mode, error) {
			return mode, nil
		})
		if shouldTerminate {
			return true, nil, ctx, nil
		}

		if mode == buildah.ModeDockerWithFuse {
			newCtx, err := InitProcessDocker(ctx, cmdData)
			if err != nil {
				return false, nil, ctx, fmt.Errorf("unable to init process docker for buildah container runtime: %s", err)
			}
			ctx = newCtx
		}

		b, err := buildah.NewBuildah(mode, buildah.BuildahOpts{})
		if err != nil {
			return false, nil, ctx, fmt.Errorf("unable to get buildah client: %s", err)
		}

		return false, container_runtime.NewBuildahRuntime(b), ctx, nil
	}

	newCtx, err := InitProcessDocker(ctx, cmdData)
	if err != nil {
		return false, nil, ctx, fmt.Errorf("unable to init process docker for docker-server container runtime: %s", err)
	}
	ctx = newCtx

	return false, container_runtime.NewDockerServerRuntime(), ctx, nil
}

func InitProcessDocker(ctx context.Context, cmdData *CmdData) (context.Context, error) {
	if err := docker.Init(ctx, *cmdData.DockerConfig, *cmdData.LogVerbose, *cmdData.LogDebug, *cmdData.Platform); err != nil {
		return ctx, fmt.Errorf("unable to init docker for buildah container runtime: %s", err)
	}

	ctxWithDockerCli, err := docker.NewContext(ctx)
	if err != nil {
		return ctx, fmt.Errorf("unable to init context for docker: %s", err)
	}
	ctx = ctxWithDockerCli

	return ctx, nil
}
