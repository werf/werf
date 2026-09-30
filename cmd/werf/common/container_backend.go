package common

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v3/pkg/buildah"
	"github.com/werf/werf/v3/pkg/buildah/thirdparty"
	"github.com/werf/werf/v3/pkg/container_backend"
	"github.com/werf/werf/v3/pkg/docker"
	"github.com/werf/werf/v3/pkg/werf"
)

func ContainerBackendProcessStartupHook() (bool, error) {
	switch {
	case strings.HasPrefix(os.Args[0], "buildah-") || strings.HasPrefix(os.Args[0], "chrootuser-") || strings.HasPrefix(os.Args[0], "storage-"):
	case os.Getenv("WERF_ORIGINAL_EXECUTABLE") == "":
		if err := os.Setenv("WERF_ORIGINAL_EXECUTABLE", os.Args[0]); err != nil {
			return false, fmt.Errorf("error setting werf original args env var: %w", err)
		}
	}

	buildahMode, _, err := GetBuildahMode()
	if err != nil {
		return false, fmt.Errorf("unable to determine buildah mode: %w", err)
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
	case "default", "auto":
		mode = buildah.ModeAuto
		var err error
		isolation, err = buildah.GetDefaultIsolation()
		if err != nil {
			return nil, nil, fmt.Errorf("unable to determine default isolation: %w", err)
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

func InitProcessContainerBackend(ctx context.Context, cmdData *CmdData, registryMirrors []string, opts InitProcessDockerOptions) (container_backend.ContainerBackend, context.Context, error) {
	buildahMode, buildahIsolation, err := GetBuildahMode()
	if err != nil {
		return nil, ctx, fmt.Errorf("unable to determine buildah mode: %w", err)
	}

	if *buildahMode != buildah.ModeDisabled {
		storageDriver, err := GetBuildahStorageDriver()
		if err != nil {
			return nil, ctx, fmt.Errorf("unable to determine buildah container backend storage driver: %w", err)
		}

		insecure := *cmdData.InsecureRegistry || *cmdData.SkipTlsVerifyRegistry

		var insecureHosts []string
		var standaloneInsecureHosts []string
		if !insecure {
			insecureHosts, err = GetInsecureRegistryHosts(ctx, cmdData, *buildahMode)
			if err != nil {
				return nil, ctx, fmt.Errorf("get insecure registry hosts: %w", err)
			}
			standaloneInsecureHosts, err = buildah.GetStandaloneInsecureRegistriesFromConfig(ctx)
			if err != nil {
				return nil, ctx, fmt.Errorf("get standalone insecure registries from containers config: %w", err)
			}
		}

		b, err := buildah.NewBuildah(*buildahMode, buildah.BuildahOpts{
			CommonBuildahOpts: buildah.CommonBuildahOpts{
				TmpDir:                       filepath.Join(werf.GetServiceDir(), "tmp", "buildah"),
				Insecure:                     insecure,
				Isolation:                    buildahIsolation,
				StorageDriver:                storageDriver,
				RegistryMirrors:              registryMirrors,
				InsecureRegistries:           insecureHosts,
				StandaloneInsecureRegistries: standaloneInsecureHosts,
			},
			NativeModeOpts: buildah.NativeModeOpts{},
		})
		if err != nil {
			return nil, ctx, fmt.Errorf("unable to get buildah client: %w", err)
		}

		return wrapContainerBackend(container_backend.NewBuildahBackend(b, container_backend.BuildahBackendOptions{TmpDir: filepath.Join(werf.GetServiceDir(), "tmp", "buildah")})), ctx, nil
	}

	newCtx, err := InitProcessDocker(ctx, cmdData, opts)
	if err != nil {
		return nil, ctx, fmt.Errorf("unable to init process docker for docker-server container backend: %w", err)
	}
	ctx = newCtx

	return wrapContainerBackend(container_backend.NewDockerServerBackend(werf.HostLocker().Locker())), ctx, nil
}

// InitProcessDockerOptions.RequireDaemon tells whether the command actually talks to the Docker
// daemon. Commands which only work with a container registry keep working against a daemon which
// is too old or not running at all; for the rest an unsupported daemon is reported right away
// instead of deep inside the first Engine API call. A daemon which is not running at all is
// tolerated even then: it is only rejected once a command really needs it.
type InitProcessDockerOptions struct {
	RequireDaemon bool
}

func InitProcessDocker(ctx context.Context, cmdData *CmdData, opts InitProcessDockerOptions) (context.Context, error) {
	if docker.IsContext(ctx) {
		if opts.RequireDaemon {
			if err := docker.CheckConnection(ctx, docker.CheckConnectionOptions{AllowDaemonUnavailable: true}); err != nil {
				return ctx, err
			}
		}
		return ctx, nil
	}

	var defaultPlatform string
	if platforms := cmdData.GetPlatform(); len(platforms) == 1 {
		defaultPlatform = platforms[0]
	}

	initOpts := docker.InitOptions{
		DockerConfigDir: *cmdData.DockerConfig,
		DefaultPlatform: defaultPlatform,
		ClaimPlatforms:  cmdData.GetPlatform(),
		Verbose:         *cmdData.LogVerbose,
		Debug:           *cmdData.LogDebug,
	}

	if err := docker.Init(ctx, initOpts); err != nil {
		return ctx, fmt.Errorf("unable to init docker: %w", err)
	}

	ctxWithDockerCli, err := docker.NewContext(ctx)
	if err != nil {
		return ctx, fmt.Errorf("unable to init context for docker: %w", err)
	}
	if opts.RequireDaemon {
		if err := docker.CheckConnection(ctxWithDockerCli, docker.CheckConnectionOptions{AllowDaemonUnavailable: true}); err != nil {
			return ctx, err
		}
	}

	return ctxWithDockerCli, nil
}
