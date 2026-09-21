package contback

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"runtime"

	. "github.com/onsi/ginkgo/v2"

	"github.com/werf/werf/v2/pkg/buildah"
	bdTypes "github.com/werf/werf/v2/pkg/buildah/thirdparty"
	"github.com/werf/werf/v2/test/pkg/thirdparty/contruntime/manifest"
)

var ErrRuntimeUnavailable = errors.New("requested runtime unavailable")

// NewContainerBackend skips the current spec when the requested backend is not
// available on this host.
func NewContainerBackend(mode string) ContainerBackend {
	switch mode {
	case "docker":
		return NewDockerBackend()
	case "native-rootless":
		SkipIfUnavailable(mode)
		return NewNativeBuildahBackend(bdTypes.IsolationOCIRootless, buildah.DefaultStorageDriver)
	case "native-chroot":
		SkipIfUnavailable(mode)
		return NewNativeBuildahBackend(bdTypes.IsolationChroot, buildah.DefaultStorageDriver)
	default:
		panic(fmt.Sprintf("unexpected buildah mode: %s", mode))
	}
}

// SkipIfUnavailable skips the current spec when the backend it needs is not
// available on this host, for specs that never touch the backend directly.
func SkipIfUnavailable(mode string) {
	switch mode {
	case "docker":
	case "native-rootless", "native-chroot", "auto", "default":
		if !buildahAvailable() {
			Skip(ErrRuntimeUnavailable.Error())
		}
	default:
		panic(fmt.Sprintf("unexpected buildah mode: %s", mode))
	}
}

func buildahAvailable() bool {
	if runtime.GOOS != "linux" {
		return false
	}

	_, err := exec.LookPath("buildah")

	return err == nil
}

type ContainerBackend interface {
	Pull(ctx context.Context, image string)
	Exec(ctx context.Context, containerName string, cmds ...string)
	Rm(ctx context.Context, containerName string)

	RunSleepingContainer(ctx context.Context, containerName, image string)
	GetImageInspect(ctx context.Context, image string) DockerImageInspect
	ExpectCmdsToSucceed(ctx context.Context, image string, cmds ...string)
}

type DockerImageInspect struct {
	Author       string
	Config       manifest.Schema2Config
	Architecture string
	Os           string
	Variant      string
	History      interface{}
}
