package buildah

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/opencontainers/runtime-spec/specs-go"

	"github.com/werf/werf/pkg/buildah/thirdparty"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
)

const (
	DefaultShmSize              = "65536k"
	DefaultSignaturePolicy      = `{"default": [{"type": "insecureAcceptAnything"}], "transports": {"docker-daemon": {"": [{"type": "insecureAcceptAnything"}]}}}`
	DefaultRegistriesConfig     = `unqualified-search-registries = ["docker.io"]`
	DefaultRuntime              = "crun"
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
	LogWriter io.Writer
}

type BuildFromDockerfileOpts struct {
	CommonOpts
	ContextTar io.Reader
	BuildArgs  map[string]string
	Target     string
}

type RunMount struct {
	Type        string
	TmpfsSize   string
	Source      string
	Destination string
}

type RunCommandOpts struct {
	CommonOpts
	Args   []string
	Mounts []specs.Mount
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
	Labels      []string
	Volumes     []string
	Expose      []string
	Envs        map[string]string
	Cmd         []string
	Entrypoint  []string
	User        string
	Workdir     string
	Healthcheck string
}

type (
	FromCommandOpts CommonOpts
	PushOpts        CommonOpts
	PullOpts        CommonOpts
	TagOpts         CommonOpts
	MountOpts       CommonOpts
	UmountOpts      CommonOpts
	RmOpts          CommonOpts
)

type Buildah interface {
	Tag(ctx context.Context, ref, newRef string, opts TagOpts) error
	Push(ctx context.Context, ref string, opts PushOpts) error
	BuildFromDockerfile(ctx context.Context, dockerfile []byte, opts BuildFromDockerfileOpts) (string, error)
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
}

type Mode string

const (
	ModeAuto           Mode = "auto"
	ModeDisabled       Mode = "disabled"
	ModeNative         Mode = "native"
	ModeDockerWithFuse Mode = "docker-with-fuse"
)

type CommonBuildahOpts struct {
	Isolation     *thirdparty.Isolation
	StorageDriver *StorageDriver
	TmpDir        string
	Insecure      bool
}

type NativeModeOpts struct {
	Platform string
}

type DockerWithFuseModeOpts struct{}

type BuildahOpts struct {
	CommonBuildahOpts
	DockerWithFuseModeOpts
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

	switch ResolveMode(mode) {
	case ModeNative:
		switch runtime.GOOS {
		case "linux":
			b, err = NewNativeBuildah(opts.CommonBuildahOpts, opts.NativeModeOpts)
			if err != nil {
				return nil, fmt.Errorf("unable to create new Buildah instance with mode %q: %w", mode, err)
			}
		default:
			panic("ModeNative can't be used on this OS")
		}
	case ModeDockerWithFuse:
		b, err = NewDockerWithFuseBuildah(opts.CommonBuildahOpts, opts.DockerWithFuseModeOpts)
		if err != nil {
			return nil, fmt.Errorf("unable to create new Buildah instance with mode %q: %w", mode, err)
		}
	default:
		return nil, fmt.Errorf("unsupported mode %q", mode)
	}

	return b, nil
}

func ProcessStartupHook(mode Mode) (bool, error) {
	switch ResolveMode(mode) {
	case ModeNative:
		return NativeProcessStartupHook(), nil
	case ModeDockerWithFuse:
		return false, nil
	default:
		return false, fmt.Errorf("unsupported mode %q", mode)
	}
}

func ResolveMode(mode Mode) Mode {
	switch mode {
	case ModeAuto:
		switch runtime.GOOS {
		case "linux":
			return ModeNative
		default:
			return ModeDockerWithFuse
		}
	default:
		return mode
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
	if util.IsInContainer() {
		return thirdparty.IsolationChroot, nil
	}
	return thirdparty.IsolationOCIRootless, nil
}

func debug() bool {
	return os.Getenv("WERF_BUILDAH_DEBUG") == "1"
}
