package common

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/werf/werf/pkg/werf"

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

		if mode == buildah.ModeDockerWithFuse {
			newCtx, err := InitProcessDocker(ctx, cmdData)
			if err != nil {
				return false, nil, ctx, fmt.Errorf("unable to init process docker for buildah container runtime: %s", err)
			}
			ctx = newCtx
		}

		buildah, err := buildah.NewBuildah(mode, buildah.BuildahOpts{CommonBuildahOpts: buildah.CommonBuildahOpts{
			TmpDir: filepath.Join(werf.GetHomeDir(), "buildah", "tmp"),
		}})
		if err != nil {
			return false, nil, ctx, fmt.Errorf("unable to get buildah client: %s", err)
		}

		return shouldTerminate, container_runtime.NewBuildahRuntime(buildah), ctx, nil
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
