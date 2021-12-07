package common

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/buildah/types"
	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/util"
)

func ContainerRuntimeProcessStartupHook() (bool, error) {
	buildahMode, _, err := GetBuildahMode()
	if err != nil {
		return false, fmt.Errorf("unable to determine buildah mode: %s", err)
	}

	switch {
	case *buildahMode != buildah.ModeDisabled:
		return buildah.ProcessStartupHook(*buildahMode)
	case strings.HasPrefix(os.Args[0], "buildah-") || strings.HasPrefix(os.Args[0], "chrootuser-") || strings.HasPrefix(os.Args[0], "storage-"):
		return buildah.ProcessStartupHook(buildah.ModeNativeRootless)
	}

	return false, nil
}

func GetBuildahMode() (*buildah.Mode, *types.Isolation, error) {
	var (
		mode      buildah.Mode
		isolation types.Isolation
	)

	modeRaw := os.Getenv("WERF_BUILDAH_MODE")
	switch modeRaw {
	case "native-rootless":
		if isInContainer, err := util.IsInContainer(); err != nil {
			return nil, nil, fmt.Errorf("unable to determine if is in container: %s", err)
		} else if isInContainer {
			return nil, nil, fmt.Errorf("native rootless mode is not available in containers: %s", err)
		}
		mode = buildah.ModeNativeRootless
		isolation = types.IsolationOCIRootless
	case "native-chroot":
		mode = buildah.ModeNativeRootless
		isolation = types.IsolationChroot
	case "docker-with-fuse":
		mode = buildah.ModeDockerWithFuse
		isolation = types.IsolationChroot
	case "default", "auto":
		mode = buildah.ModeAuto
		var err error
		isolation, err = buildah.GetDefaultIsolation()
		if err != nil {
			return nil, nil, fmt.Errorf("unable to determine default isolation: %s", err)
		}
	case "docker", "":
		mode = buildah.ModeDisabled
	default:
		return nil, nil, fmt.Errorf("unexpected mode specified: %s", modeRaw)
	}

	return &mode, &isolation, nil
}

func GetBuildahStorageDriver() (*buildah.StorageDriver, error) {
	storageDriverRaw := os.Getenv("WERF_BUILDAH_STORAGE_DRIVER")
	var storageDriver buildah.StorageDriver
	switch storageDriverRaw {
	case string(buildah.StorageDriverOverlay), string(buildah.StorageDriverVFS):
		storageDriver = buildah.StorageDriver(storageDriverRaw)
	case "default", "auto", "":
		storageDriver = buildah.DefaultStorageDriver
	default:
		return nil, fmt.Errorf("unexpected driver specified: %s", storageDriverRaw)
	}
	return &storageDriver, nil
}

func wrapContainerRuntime(containerRuntime container_runtime.ContainerRuntime) container_runtime.ContainerRuntime {
	if os.Getenv("WERF_PERF_TEST_CONTAINER_RUNTIME") == "1" {
		return container_runtime.NewPerfCheckContainerRuntime(containerRuntime)
	}
	return containerRuntime
}

func InitProcessContainerRuntime(ctx context.Context, cmdData *CmdData) (container_runtime.ContainerRuntime, context.Context, error) {
	buildahMode, buildahIsolation, err := GetBuildahMode()
	if err != nil {
		return nil, ctx, fmt.Errorf("unable to determine buildah mode: %s", err)
	}

	if *buildahMode != buildah.ModeDisabled {
		resolvedMode := buildah.ResolveMode(*buildahMode)
		if resolvedMode == buildah.ModeDockerWithFuse {
			newCtx, err := InitProcessDocker(ctx, cmdData)
			if err != nil {
				return nil, ctx, fmt.Errorf("unable to init process docker for buildah container runtime: %s", err)
			}
			ctx = newCtx
		}

		storageDriver, err := GetBuildahStorageDriver()
		if err != nil {
			return nil, ctx, fmt.Errorf("unable to determine buildah container runtime storage driver: %s", err)
		}

		insecure := *cmdData.InsecureRegistry || *cmdData.SkipTlsVerifyRegistry
		b, err := buildah.NewBuildah(resolvedMode, buildah.BuildahOpts{
			CommonBuildahOpts: buildah.CommonBuildahOpts{
				Insecure:      insecure,
				Isolation:     buildahIsolation,
				StorageDriver: storageDriver,
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
