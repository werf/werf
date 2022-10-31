package container_backend

import (
	"context"
	"sync"

	"github.com/docker/docker/api/types"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/docker"
)

var (
	runningDockerContainers []*runningDockerContainer
	mu                      sync.Mutex
)

type runningDockerContainer struct {
	Name string
	Ctx  context.Context
}

func RegisterRunningContainer(name string, ctx context.Context) {
	mu.Lock()
	defer mu.Unlock()

	runningDockerContainers = append(runningDockerContainers, &runningDockerContainer{
		Name: name,
		Ctx:  ctx,
	})
}

func UnregisterRunningContainer(name string) {
	mu.Lock()
	defer mu.Unlock()

	var res []*runningDockerContainer
	for _, cont := range runningDockerContainers {
		if cont.Name != name {
			res = append(res, cont)
		}
	}
	runningDockerContainers = res
}

func TerminateRunningDockerContainers() error {
	mu.Lock()
	defer mu.Unlock()

	for _, container := range runningDockerContainers {
		logboek.Context(container.Ctx).Info().LogF("Removing container %q...\n", container.Name)

		err := docker.ContainerRemove(container.Ctx, container.Name, types.ContainerRemoveOptions{RemoveVolumes: true, Force: true})
		if err != nil {
			logboek.Context(container.Ctx).Error().LogF("WARNING: Unable to remove container %q: %w\n", container.Name, err.Error())
		}
	}

	return nil
}
