package common

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/docker"
)

func ContainerRuntimeProcessStartupHook() (bool, error) {
	buildahMode := GetContainerRuntimeBuildahMode()

	switch {
	case buildahMode != "":
		return buildah.ProcessStartupHook(buildahMode)
	case strings.HasPrefix(os.Args[0], "buildah-") || strings.HasPrefix(os.Args[0], "chrootuser-") || strings.HasPrefix(os.Args[0], "storage-"):
		return buildah.ProcessStartupHook("native-rootless")
	}

	return false, nil
}

func GetContainerRuntimeBuildahMode() buildah.Mode {
	return buildah.Mode(os.Getenv("WERF_CONTAINER_RUNTIME_BUILDAH"))
}

func wrapContainerRuntime(containerRuntime container_runtime.ContainerRuntime) container_runtime.ContainerRuntime {
	if os.Getenv("WERF_PERF_TEST_CONTAINER_RUNTIME") == "1" {
		return container_runtime.NewPerfCheckContainerRuntime(containerRuntime)
	}
	return containerRuntime
}

func InitProcessContainerRuntime(ctx context.Context, cmdData *CmdData) (container_runtime.ContainerRuntime, context.Context, error) {
	buildahMode := GetContainerRuntimeBuildahMode()
	if buildahMode != "" {
		resolvedMode := buildah.ResolveMode(buildahMode)
		if resolvedMode == buildah.ModeDockerWithFuse {
			newCtx, err := InitProcessDocker(ctx, cmdData)
			if err != nil {
				return nil, ctx, fmt.Errorf("unable to init process docker for buildah container runtime: %s", err)
			}
			ctx = newCtx
		}

		insecure := *cmdData.InsecureRegistry || *cmdData.SkipTlsVerifyRegistry
		b, err := buildah.NewBuildah(resolvedMode, buildah.BuildahOpts{
			CommonBuildahOpts: buildah.CommonBuildahOpts{
				Insecure: insecure,
			},
		})
		if err != nil {
			return nil, ctx, fmt.Errorf("unable to get buildah client: %s", err)
		}

		return wrapContainerRuntime(container_runtime.NewBuildahRuntime(b)), ctx, nil
	}

	newCtx, err := InitProcessDocker(ctx, cmdData)
	if err != nil {
		return nil, ctx, fmt.Errorf("unable to init process docker for docker-server container runtime: %s", err)
	}
	ctx = newCtx

	return wrapContainerRuntime(container_runtime.NewDockerServerRuntime()), ctx, nil
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
