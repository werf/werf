package buildah

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/werf/werf/pkg/buildah/types"
	"github.com/werf/werf/pkg/werf"
)

const (
	DefaultShmSize              = "65536k"
	BuildahImage                = "ghcr.io/werf/buildah:v1.22.3-1"
	BuildahStorageContainerName = "werf-buildah-storage"
)

type CommonOpts struct {
	LogWriter io.Writer
}

type BuildFromDockerfileOpts struct {
	CommonOpts
	ContextTar io.Reader
}

type RunCommandOpts struct {
	CommonOpts
	BuildArgs []string
}

type FromCommandOpts struct {
	CommonOpts
}

type PullOpts struct {
	CommonOpts
}

type PushOpts struct {
	CommonOpts
}

type TagOpts struct {
	CommonOpts
}

type RmiOpts struct {
	CommonOpts
}

type Buildah interface {
	Tag(ctx context.Context, ref, newRef string, opts TagOpts) error
	Push(ctx context.Context, ref string, opts PushOpts) error
	BuildFromDockerfile(ctx context.Context, dockerfile []byte, opts BuildFromDockerfileOpts) (string, error)
	RunCommand(ctx context.Context, container string, command []string, opts RunCommandOpts) error
	FromCommand(ctx context.Context, container string, image string, opts FromCommandOpts) error
	Pull(ctx context.Context, ref string, opts PullOpts) error
	Inspect(ctx context.Context, ref string) (*types.BuilderInfo, error)
	Rmi(ctx context.Context, ref string, opts RmiOpts) error
}

type Mode string

const (
	ModeAuto           Mode = "auto"
	ModeNativeRootless Mode = "native-rootless"
	ModeDockerWithFuse Mode = "docker-with-fuse"
)

func InitProcess(initModeFunc func() (Mode, error)) (bool, Mode, error) {
	if v := os.Getenv("_BUILDAH_PROCESS_INIT_MODE"); v != "" {
		mode := Mode(v)
		shouldTerminate, err := doInitProcess(mode)
		return shouldTerminate, mode, err
	}

	mode, err := initModeFunc()
	if err != nil {
		return false, "", fmt.Errorf("unable to init buildah mode: %s", err)
	}
	os.Setenv("_BUILDAH_PROCESS_INIT_MODE", string(mode))

	shouldTerminate, err := doInitProcess(mode)
	return shouldTerminate, mode, err
}

func doInitProcess(mode Mode) (bool, error) {
	switch resolveMode(mode) {
	case ModeNativeRootless:
		return InitNativeRootlessProcess()
	case ModeDockerWithFuse:
		return false, nil
	default:
		return false, fmt.Errorf("unsupported mode %q", mode)
	}
}

type CommonBuildahOpts struct {
	TmpDir string
}

type NativeRootlessModeOpts struct{}

type DockerWithFuseModeOpts struct{}

type BuildahOpts struct {
	CommonBuildahOpts
	DockerWithFuseModeOpts
	NativeRootlessModeOpts
}

func NewBuildah(mode Mode, opts BuildahOpts) (b Buildah, err error) {
	if opts.CommonBuildahOpts.TmpDir == "" {
		opts.CommonBuildahOpts.TmpDir = filepath.Join(werf.GetHomeDir(), "buildah", "tmp")
	}

	switch resolveMode(mode) {
	case ModeNativeRootless:
		switch runtime.GOOS {
		case "linux":
			b, err = NewNativeRootlessBuildah(opts.CommonBuildahOpts, opts.NativeRootlessModeOpts)
			if err != nil {
				return nil, fmt.Errorf("unable to create new Buildah instance with mode %q: %s", mode, err)
			}
		default:
			panic("ModeNativeRootless can't be used on this OS")
		}
	case ModeDockerWithFuse:
		b, err = NewDockerWithFuseBuildah(opts.CommonBuildahOpts, opts.DockerWithFuseModeOpts)
		if err != nil {
			return nil, fmt.Errorf("unable to create new Buildah instance with mode %q: %s", mode, err)
		}
	default:
		return nil, fmt.Errorf("unsupported mode %q", mode)
	}

	return b, nil
}

func resolveMode(mode Mode) Mode {
	switch mode {
	case ModeAuto:
		switch runtime.GOOS {
		case "linux":
			return ModeNativeRootless
		default:
			return ModeDockerWithFuse
		}
	default:
		return mode
	}
}

func debug() bool {
	return os.Getenv("WERF_BUILDAH_DEBUG") == "1"
}
