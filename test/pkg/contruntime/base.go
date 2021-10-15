package contruntime

import (
	"github.com/google/uuid"
	"github.com/werf/werf/test/pkg/thirdparty/contruntime/manifest"
)

type BuildahInspect struct {
	Docker struct {
		Config manifest.Schema2Config `json:"config"`
	} `json:"Docker"`
}

type BaseContainerRuntime struct{}

func expectCmdsToSucceed(r ContainerRuntime, image string, cmds ...string) {
	containerName := uuid.New().String()
	r.RunSleepingContainer(containerName, image)
	r.Exec(containerName, cmds...)
	r.Rm(containerName)
}
