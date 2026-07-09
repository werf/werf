package common

import (
	"context"
	"fmt"
	"os"

	"github.com/werf/werf/v2/pkg/buildkit"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/docker"
	"github.com/werf/werf/v2/pkg/storage"
)

func wrapContainerBackend(containerBackend container_backend.ContainerBackend) container_backend.ContainerBackend {
	if os.Getenv("WERF_PERF_TEST_CONTAINER_RUNTIME") == "1" {
		return container_backend.NewPerfCheckContainerBackend(containerBackend)
	}
	return containerBackend
}

func InitProcessContainerBackend(ctx context.Context, cmdData *CmdData, registryMirrors []string) (container_backend.ContainerBackend, context.Context, error) {
	var resolveHostOptions buildkit.ResolveHostOptions
	if cmdData.Repo != nil && cmdData.Repo.Address != nil && *cmdData.Repo.Address != "" && *cmdData.Repo.Address != storage.LocalStorageAddress {
		if *cmdData.InsecureRegistry {
			resolveHostOptions.InsecureRegistryAddresses = append(resolveHostOptions.InsecureRegistryAddresses, *cmdData.Repo.Address)
		}
		if *cmdData.SkipTlsVerifyRegistry {
			resolveHostOptions.SkipTLSVerifyRegistryAddresses = append(resolveHostOptions.SkipTLSVerifyRegistryAddresses, *cmdData.Repo.Address)
		}
	}

	buildkitHost, err := buildkit.ResolveHost(ctx, resolveHostOptions)
	if err != nil {
		return nil, ctx, err
	}

	var defaultPlatform string
	if platforms := cmdData.GetPlatform(); len(platforms) == 1 {
		defaultPlatform = platforms[0]
	}

	return wrapContainerBackend(container_backend.NewBuildkitBackend(buildkitHost, container_backend.BuildkitBackendOptions{
		DefaultPlatform:       defaultPlatform,
		DockerConfigDir:       *cmdData.DockerConfig,
		InsecureRegistry:      *cmdData.InsecureRegistry,
		SkipTLSVerifyRegistry: *cmdData.SkipTlsVerifyRegistry,
	})), ctx, nil
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
