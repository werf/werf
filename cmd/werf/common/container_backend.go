package common

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/buildah/thirdparty"
	"github.com/werf/werf/pkg/container_backend"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/util"
)

func ContainerBackendProcessStartupHook() (bool, error) {
	buildahMode, _, err := GetBuildahMode()
	if err != nil {
		return false, fmt.Errorf("unable to determine buildah mode: %s", err)
	}

	switch {
	case *buildahMode != buildah.ModeDisabled:
		return buildah.ProcessStartupHook(*buildahMode)
	case strings.HasPrefix(os.Args[0], "buildah-") || strings.HasPrefix(os.Args[0], "chrootuser-") || strings.HasPrefix(os.Args[0], "storage-"):
		return buildah.ProcessStartupHook(buildah.ModeNative)
	}

	return false, nil
}

func GetBuildahMode() (*buildah.Mode, *thirdparty.Isolation, error) {
	var (
		mode      buildah.Mode
		isolation thirdparty.Isolation
	)

	modeRaw := os.Getenv("WERF_BUILDAH_MODE")
	switch modeRaw {
	case "native-rootless":
		if util.IsInContainer() {
			return nil, nil, fmt.Errorf("native rootless mode is not available in containers")
		}
		mode = buildah.ModeNative
		isolation = thirdparty.IsolationOCIRootless
	case "native-chroot":
		mode = buildah.ModeNative
		isolation = thirdparty.IsolationChroot
	case "docker-with-fuse":
		mode = buildah.ModeDockerWithFuse
		isolation = thirdparty.IsolationChroot
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

func wrapContainerBackend(containerBackend container_backend.ContainerBackend) container_backend.ContainerBackend {
	if os.Getenv("WERF_PERF_TEST_CONTAINER_RUNTIME") == "1" {
		return container_backend.NewPerfCheckContainerBackend(containerBackend)
	}
	return containerBackend
}

func InitProcessContainerBackend(ctx context.Context, cmdData *CmdData) (container_backend.ContainerBackend, context.Context, error) {
	buildahMode, buildahIsolation, err := GetBuildahMode()
	if err != nil {
		return nil, ctx, fmt.Errorf("unable to determine buildah mode: %s", err)
	}

	if *buildahMode != buildah.ModeDisabled {
		resolvedMode := buildah.ResolveMode(*buildahMode)
		if resolvedMode == buildah.ModeDockerWithFuse {
			newCtx, err := InitProcessDocker(ctx, cmdData)
			if err != nil {
				return nil, ctx, fmt.Errorf("unable to init process docker for buildah container backend: %s", err)
			}
			ctx = newCtx
		}

		storageDriver, err := GetBuildahStorageDriver()
		if err != nil {
			return nil, ctx, fmt.Errorf("unable to determine buildah container backend storage driver: %s", err)
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

		return wrapContainerBackend(container_backend.NewBuildahBackend(b)), ctx, nil
	}

	newCtx, err := InitProcessDocker(ctx, cmdData)
	if err != nil {
		return nil, ctx, fmt.Errorf("unable to init process docker for docker-server container backend: %s", err)
	}
	ctx = newCtx

	return wrapContainerBackend(container_backend.NewDockerServerBackend()), ctx, nil
}

func InitProcessDocker(ctx context.Context, cmdData *CmdData) (context.Context, error) {
	if err := docker.Init(ctx, *cmdData.DockerConfig, *cmdData.LogVerbose, *cmdData.LogDebug, *cmdData.Platform); err != nil {
		return ctx, fmt.Errorf("unable to init docker for buildah container backend: %s", err)
	}

	ctxWithDockerCli, err := docker.NewContext(ctx)
	if err != nil {
		return ctx, fmt.Errorf("unable to init context for docker: %s", err)
	}
	ctx = ctxWithDockerCli

	return ctx, nil
}
