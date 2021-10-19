package contruntime

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/werf/werf/test/pkg/thirdparty/contruntime/manifest"
)

var RuntimeUnavailError = errors.New("requested runtime unavailable")

func NewContainerRuntime(name string) (ContainerRuntime, error) {
	switch name {
	case "docker":
		return NewDockerRuntime(), nil
	case "native-rootless-buildah":
		if runtime.GOOS != "linux" {
			return nil, RuntimeUnavailError
		}
		return NewNativeRootlessBuildahRuntime(), nil
	case "docker-with-fuse-buildah":
		return NewDockerWithFuseBuildahRuntime(), nil
	default:
		panic(fmt.Sprint("unexpected name for container runtime: ", name))
	}
}

type ContainerRuntime interface {
	Pull(image string)
	Exec(containerName string, cmds ...string)
	Rm(containerName string)

	RunSleepingContainer(containerName, image string)
	GetImageInspectConfig(image string) (config manifest.Schema2Config)
	ExpectCmdsToSucceed(image string, cmds ...string)
}
