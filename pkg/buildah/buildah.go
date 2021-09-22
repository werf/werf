package buildah

import (
	"context"
	"io"

	"github.com/docker/docker/pkg/reexec"

	"github.com/containers/storage/pkg/unshare"
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

type Buildah interface {
	BuildFromDockerfile(ctx context.Context, dockerfile []byte, opts BuildFromDockerfileOpts) (string, error)
	RunCommand(ctx context.Context, container string, command []string, opts RunCommandOpts) error
}

type BuildahMode string

var (
	NativeRootless BuildahMode = "NativeRootless"
	DockerWithFuse BuildahMode = "DockerWithFuse"
)

type NewBuildahOpts struct {
	Mode BuildahMode
}

func NewBuildah(mode BuildahMode) (Buildah, error) {
	switch mode {
	case "":
		// TODO: auto select based on OS
	case NativeRootless:
		// TODO: validate selected mode with OS
	case DockerWithFuse:
		// TODO: validate selected mode with OS
	}

	switch mode {
	case NativeRootless:
		return NewNativeRootlessBuildah()
	case DockerWithFuse:
		return NewDockerWithFuseBuildah()
	default:
		panic("unexpected")
	}
}

func InitProcess(mode BuildahMode) error {
	switch mode {
	case NativeRootless:
		if reexec.Init() {
			return nil
		}

		unshare.MaybeReexecUsingUserNamespace(false)

		return nil
	case DockerWithFuse:
		return nil
	default:
		panic("unexpected")
	}
}
