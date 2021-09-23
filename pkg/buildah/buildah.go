package buildah

import (
	"context"
	"fmt"
	"io"
	"runtime"
)

const (
	DefaultShmSize = "65536k"
	BuildahImage   = "quay.io/buildah/stable:v1.22.3"
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

type Buildah interface {
	BuildFromDockerfile(ctx context.Context, dockerfile []byte, opts BuildFromDockerfileOpts) (string, error)
	RunCommand(ctx context.Context, container string, command []string, opts RunCommandOpts) error
	FromCommand(ctx context.Context, container string, image string, opts FromCommandOpts) error
}

type Mode int

const (
	ModeAuto Mode = iota
	ModeNativeRootless
	ModeDockerWithFuse
)

func NewBuildah(mode Mode) (b Buildah, err error) {
	switch mode {
	case ModeAuto:
		switch runtime.GOOS {
		case "linux":
			return NewBuildah(ModeNativeRootless)
		default:
			return NewBuildah(ModeDockerWithFuse)
		}
	case ModeNativeRootless:
		switch runtime.GOOS {
		case "linux":
			b, err = NewNativeRootlessBuildah()
			if err != nil {
				return nil, fmt.Errorf("unable to create new Buildah instance with mode %d: %s", mode, err)
			}
		default:
			panic("ModeNativeRootless can't be used on this OS")
		}
	case ModeDockerWithFuse:
		b, err = NewDockerWithFuseBuildah()
		if err != nil {
			return nil, fmt.Errorf("unable to create new Buildah instance with mode %d: %s", mode, err)
		}
	default:
		panic(fmt.Sprintf("unexpected Mode: %d", mode))
	}

	return b, nil
}
