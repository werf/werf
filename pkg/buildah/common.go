package buildah

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/opencontainers/runtime-spec/specs-go"

	"github.com/werf/werf/pkg/buildah/thirdparty"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
)

const (
	DefaultShmSize          = "65536k"
	DefaultContainersConfig = `
[network]
default_rootless_network_cmd="slirp4netns"
[engine]
# Prefer runc over crun since old versions of crun (including one shipped in Ubuntu 22.04) cause
# "unknown version specified" error
runtime="runc"
`
	DefaultSignaturePolicy      = `{"default": [{"type": "insecureAcceptAnything"}], "transports": {"docker-daemon": {"": [{"type": "insecureAcceptAnything"}]}}}`
	DefaultRegistriesConfig     = `unqualified-search-registries = ["docker.io"]`
	BuildahImage                = "registry.werf.io/werf/buildah:v1.22.3-1"
	BuildahStorageContainerName = "werf-buildah-storage"

	DefaultStorageDriver StorageDriver = StorageDriverOverlay
)

type StorageDriver string

const (
	StorageDriverOverlay StorageDriver = "overlay"
	StorageDriverVFS     StorageDriver = "vfs"
)

type CommonOpts struct {
	TargetPlatform string
	LogWriter      io.Writer
}

type BuildFromDockerfileOpts struct {
	CommonOpts

	ContextDir string
	BuildArgs  map[string]string
	Target     string
	Labels     []string
}

type RunMount struct {
	Type        string
	TmpfsSize   string
	Source      string
	Destination string
}

type RunCommandOpts struct {
	CommonOpts

	ContextDir       string
	PrependShell     bool
	Shell            []string
	AddCapabilities  []string
	DropCapabilities []string
	NetworkType      string
	WorkingDir       string
	User             string
	Envs             []string
	// Mounts as allowed to be passed from command line.
	GlobalMounts []*specs.Mount
	// Mounts as allowed in Dockerfile RUN --mount option. Have more restrictions than GlobalMounts (e.g. Source of bind-mount can't be outside of ContextDir or container root).
	RunMounts []*instructions.Mount
}

type RmiOpts struct {
	CommonOpts

	Force bool
}

type CommitOpts struct {
	CommonOpts

	Image string
}

type ConfigOpts struct {
	CommonOpts

	Labels                 []string
	Volumes                []string
	Expose                 []string
	Envs                   map[string]string
	Cmd                    []string
	CmdPrependShell        bool
	Entrypoint             []string
	EntrypointPrependShell bool
	User                   string
	Workdir                string
	Healthcheck            *thirdparty.BuildahHealthConfig
	OnBuild                string
	StopSignal             string
	Shell                  []string
	Maintainer             string
}

type CopyOpts struct {
	CommonOpts

	Chown   string
	Chmod   string
	Ignores []string
}

type AddOpts struct {
	CommonOpts

	ContextDir string
	Chown      string
	Chmod      string
	Ignores    []string
}

type ImagesOptions struct {
	CommitOpts
	Names   []string
	Filters []util.Pair[string, string]
}

type ContainersOptions struct {
	CommitOpts
	Filters []image.ContainerFilter
}

type (
	FromCommandOpts       CommonOpts
	BuildFromCommandsOpts CommonOpts
	PushOpts              CommonOpts
	PullOpts              CommonOpts
	TagOpts               CommonOpts
	MountOpts             CommonOpts
	UmountOpts            CommonOpts
	RmOpts                CommonOpts
)

type Buildah interface {
	GetDefaultPlatform() string
	GetRuntimePlatform() string
	Tag(ctx context.Context, ref, newRef string, opts TagOpts) error
	Push(ctx context.Context, ref string, opts PushOpts) error
	BuildFromDockerfile(ctx context.Context, dockerfile string, opts BuildFromDockerfileOpts) (string, error)
	RunCommand(ctx context.Context, container string, command []string, opts RunCommandOpts) error
	FromCommand(ctx context.Context, container, image string, opts FromCommandOpts) (string, error)
	Pull(ctx context.Context, ref string, opts PullOpts) error
	Inspect(ctx context.Context, ref string) (*thirdparty.BuilderInfo, error)
	Rm(ctx context.Context, ref string, opts RmOpts) error
	Rmi(ctx context.Context, ref string, opts RmiOpts) error
	Mount(ctx context.Context, container string, opts MountOpts) (string, error)
	Umount(ctx context.Context, container string, opts UmountOpts) error
	Commit(ctx context.Context, container string, opts CommitOpts) (string, error)
	Config(ctx context.Context, container string, opts ConfigOpts) error
	Copy(ctx context.Context, container, contextDir string, src []string, dst string, opts CopyOpts) error
	Add(ctx context.Context, container string, src []string, dst string, opts AddOpts) error
	Images(ctx context.Context, opts ImagesOptions) (image.ImagesList, error)
	Containers(ctx context.Context, opts ContainersOptions) (image.ContainerList, error)
}

type Mode string

const (
	ModeAuto     Mode = "auto"
	ModeDisabled Mode = "disabled"
	ModeNative   Mode = "native"
)

type CommonBuildahOpts struct {
	Isolation     *thirdparty.Isolation
	StorageDriver *StorageDriver
	TmpDir        string
	Insecure      bool
}

type NativeModeOpts struct {
	DefaultPlatform string
}

type BuildahOpts struct {
	CommonBuildahOpts
	NativeModeOpts
}

func NewBuildah(mode Mode, opts BuildahOpts) (b Buildah, err error) {
	if opts.CommonBuildahOpts.Isolation == nil {
		defIsolation, err := GetDefaultIsolation()
		if err != nil {
			return b, fmt.Errorf("unable to determine default isolation: %w", err)
		}
		opts.CommonBuildahOpts.Isolation = &defIsolation
	}

	if opts.CommonBuildahOpts.StorageDriver == nil {
		defStorageDriver := DefaultStorageDriver
		opts.CommonBuildahOpts.StorageDriver = &defStorageDriver
	}

	if opts.CommonBuildahOpts.TmpDir == "" {
		opts.CommonBuildahOpts.TmpDir = filepath.Join(werf.GetHomeDir(), "buildah", "tmp")
	}

	switch mode {
	case ModeNative, ModeAuto:
		switch runtime.GOOS {
		case "linux":
			b, err = NewNativeBuildah(opts.CommonBuildahOpts, opts.NativeModeOpts)
			if err != nil {
				return nil, fmt.Errorf("unable to create new Buildah instance with mode %q: %w", mode, err)
			}
		default:
			panic(fmt.Sprintf("Mode %q can't be used on this OS", mode))
		}
	default:
		return nil, fmt.Errorf("unsupported mode %q", mode)
	}

	return b, nil
}

func ProcessStartupHook(mode Mode) (bool, error) {
	switch mode {
	case ModeNative, ModeAuto:
		return NativeProcessStartupHook(), nil
	default:
		return false, fmt.Errorf("unsupported mode %q", mode)
	}
}

func GetFuseOverlayfsOptions() ([]string, error) {
	fuseOverlayBinPath, err := exec.LookPath("fuse-overlayfs")
	if err != nil {
		return nil, fmt.Errorf("\"fuse-overlayfs\" binary not found in PATH: %w", err)
	}

	result := []string{fmt.Sprintf("overlay.mount_program=%s", fuseOverlayBinPath)}

	if util.IsInContainer() {
		result = append(result, fmt.Sprintf("overlay.mountopt=%s", "nodev,fsync=0"))
	}

	return result, nil
}

func GetDefaultIsolation() (thirdparty.Isolation, error) {
	return thirdparty.IsolationChroot, nil
}

// TODO used for linux only, hide under the build flag
//
//nolint:nolintlint,unused
func debug() bool {
	return os.Getenv("WERF_BUILDAH_DEBUG") == "1"
}

type StoreOptions struct {
	GraphDriverName    string
	GraphDriverOptions []string
}

func GetBasicBuildahCliArgs(driver StorageDriver) ([]string, error) {
	var result []string

	cliStoreOpts, err := newBuildahCliStoreOptions(driver)
	if err != nil {
		return result, fmt.Errorf("unable to get buildah cli store options: %w", err)
	}

	if cliStoreOpts.GraphDriverName != "" {
		result = append(result, "--storage-driver", cliStoreOpts.GraphDriverName)
	}

	if len(cliStoreOpts.GraphDriverOptions) > 0 {
		result = append(result, "--storage-opt", strings.Join(cliStoreOpts.GraphDriverOptions, ","))
	}

	return result, nil
}

func newBuildahCliStoreOptions(driver StorageDriver) (*StoreOptions, error) {
	var graphDriverOptions []string
	if driver == StorageDriverOverlay {
		fuseOpts, err := GetFuseOverlayfsOptions()
		if err != nil {
			return nil, fmt.Errorf("unable to get overlay options: %w", err)
		}
		graphDriverOptions = append(graphDriverOptions, fuseOpts...)
	}

	return &StoreOptions{
		GraphDriverName:    string(driver),
		GraphDriverOptions: graphDriverOptions,
	}, nil
}
