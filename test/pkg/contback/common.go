package contback

import (
	"context"

	"github.com/werf/werf/v2/test/pkg/thirdparty/contruntime/manifest"
)

func NewContainerBackend() ContainerBackend {
	// buildkit-built images live in the test repo registry and are inspected via docker.
	return NewDockerBackend()
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
