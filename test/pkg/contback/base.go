package contback

import (
	"github.com/google/uuid"

	"github.com/werf/werf/pkg/buildah/thirdparty"
)

type BaseContainerBackend struct {
	CommonCliArgs []string
	Isolation     thirdparty.Isolation
}

func expectCmdsToSucceed(r ContainerBackend, image string, cmds ...string) {
	containerName := uuid.New().String()
	r.RunSleepingContainer(containerName, image)
	r.Exec(containerName, cmds...)
	r.Rm(containerName)
}
