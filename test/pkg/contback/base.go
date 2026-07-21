package contback

import (
	"context"

	"github.com/google/uuid"
)

type BaseContainerBackend struct {
	CommonCliArgs []string
}

func expectCmdsToSucceed(ctx context.Context, r ContainerBackend, image string, cmds ...string) {
	containerName := uuid.New().String()
	r.RunSleepingContainer(ctx, containerName, image)
	r.Exec(ctx, containerName, cmds...)
	r.Rm(ctx, containerName)
}
