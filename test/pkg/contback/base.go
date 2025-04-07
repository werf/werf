package contback

import (
	"context"

	"github.com/google/uuid"

	"github.com/werf/werf/v2/pkg/buildah/thirdparty"
)

type BaseContainerBackend struct {
	CommonCliArgs []string
	Isolation     thirdparty.Isolation
}

func expectCmdsToSucceed(ctx context.Context, r ContainerBackend, image string, cmds ...string) {
	containerName := uuid.New().String()
	r.RunSleepingContainer(ctx, containerName, image)
	r.Exec(ctx, containerName, cmds...)
	r.Rm(ctx, containerName)
}
