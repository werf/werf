package contback

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/werf/werf/pkg/buildah"
	bdTypes "github.com/werf/werf/pkg/buildah/thirdparty"
	"github.com/werf/werf/test/pkg/thirdparty/contruntime/manifest"
)

var ErrRuntimeUnavailable = errors.New("requested runtime unavailable")

func NewContainerBackend(buildahMode string) (ContainerBackend, error) {
	switch buildahMode {
	case "docker":
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
	case "docker-with-fuse":
		return NewDockerWithFuseBuildahBackend(bdTypes.IsolationChroot, buildah.DefaultStorageDriver), nil
	default:
		panic(fmt.Sprintf("unexpected buildah mode: %s", buildahMode))
	}
}

type ContainerBackend interface {
	Pull(image string)
	Exec(containerName string, cmds ...string)
	Rm(containerName string)

	RunSleepingContainer(containerName, image string)
	GetImageInspectConfig(image string) (config manifest.Schema2Config)
	ExpectCmdsToSucceed(image string, cmds ...string)
}
