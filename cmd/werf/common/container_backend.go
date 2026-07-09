package common

import (
	"context"
	"fmt"
	"os"

	"github.com/werf/werf/v2/pkg/buildkit"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/docker"
	"github.com/werf/werf/v2/pkg/werf"
)

func wrapContainerBackend(containerBackend container_backend.ContainerBackend) container_backend.ContainerBackend {
	if os.Getenv("WERF_PERF_TEST_CONTAINER_RUNTIME") == "1" {
		return container_backend.NewPerfCheckContainerBackend(containerBackend)
	}
	return containerBackend
}

func InitProcessContainerBackend(ctx context.Context, cmdData *CmdData, registryMirrors []string) (container_backend.ContainerBackend, context.Context, error) {
	if buildkitHost := buildkit.HostFromEnv(); buildkitHost != "" {
		var defaultPlatform string
		if platforms := cmdData.GetPlatform(); len(platforms) == 1 {
			defaultPlatform = platforms[0]
		}

		return wrapContainerBackend(container_backend.NewBuildkitBackend(buildkitHost, container_backend.BuildkitBackendOptions{
			DefaultPlatform: defaultPlatform,
			DockerConfigDir: *cmdData.DockerConfig,
		})), ctx, nil
	}

	newCtx, err := InitProcessDocker(ctx, cmdData)
	if err != nil {
		return nil, ctx, fmt.Errorf("unable to init process docker for docker-server container backend: %w", err)
	}
	ctx = newCtx

	return wrapContainerBackend(container_backend.NewDockerServerBackend(werf.HostLocker().Locker())), ctx, nil
}

func InitProcessDocker(ctx context.Context, cmdData *CmdData) (context.Context, error) {
	if docker.IsContext(ctx) {
		return ctx, nil
	}

	var defaultPlatform string
	if platforms := cmdData.GetPlatform(); len(platforms) == 1 {
		defaultPlatform = platforms[0]
	}

	opts := docker.InitOptions{
		DockerConfigDir: *cmdData.DockerConfig,
		DefaultPlatform: defaultPlatform,
		ClaimPlatforms:  cmdData.GetPlatform(),
		Verbose:         *cmdData.LogVerbose,
		Debug:           *cmdData.LogDebug,
	}

	if err := docker.Init(ctx, opts); err != nil {
		return ctx, fmt.Errorf("unable to init docker: %w", err)
	}

	ctxWithDockerCli, err := docker.NewContext(ctx)
	if err != nil {
		return ctx, fmt.Errorf("unable to init context for docker: %w", err)
	}
	ctx = ctxWithDockerCli

	return ctx, nil
}
