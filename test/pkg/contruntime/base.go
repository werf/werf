package contruntime

import (
	"github.com/google/uuid"

	"github.com/werf/werf/pkg/buildah/types"
)

type BaseContainerRuntime struct {
	CommonCliArgs []string
	Isolation     types.Isolation
}

func expectCmdsToSucceed(r ContainerRuntime, image string, cmds ...string) {
	containerName := uuid.New().String()
	r.RunSleepingContainer(containerName, image)
	r.Exec(containerName, cmds...)
	r.Rm(containerName)
}
