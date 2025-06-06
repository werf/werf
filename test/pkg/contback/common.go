package contback

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"runtime"

	"github.com/werf/werf/v2/pkg/buildah"
	bdTypes "github.com/werf/werf/v2/pkg/buildah/thirdparty"
	"github.com/werf/werf/v2/test/pkg/thirdparty/contruntime/manifest"
)

var ErrRuntimeUnavailable = errors.New("requested runtime unavailable")

func NewContainerBackend(mode string) (ContainerBackend, error) {
	switch mode {
	case "docker", "vanilla-docker", "buildkit-docker":
		return NewDockerBackend(), nil
	case "native-rootless":
		if runtime.GOOS != "linux" {
			return nil, ErrRuntimeUnavailable
		}
		return NewNativeBuildahBackend(bdTypes.IsolationOCIRootless, buildah.DefaultStorageDriver), nil
	case "native-chroot":
		if runtime.GOOS != "linux" {
			return nil, ErrRuntimeUnavailable
		}
		return NewNativeBuildahBackend(bdTypes.IsolationChroot, buildah.DefaultStorageDriver), nil
	default:
		panic(fmt.Sprintf("unexpected buildah mode: %s", mode))
	}
}

type ContainerBackend interface {
	Pull(ctx context.Context, image string)
	Exec(ctx context.Context, containerName string, cmds ...string)
	Rm(ctx context.Context, containerName string)
	DumpImage(ctx context.Context, image string) *bytes.Reader

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
