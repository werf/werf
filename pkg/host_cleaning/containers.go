package host_cleaning

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"

	"github.com/flant/logboek"

	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/image"
)

func werfContainersFlushByFilterSet(filterSet filters.Args, options CommonOptions) error {
	containers, err := werfContainersByFilterSet(filterSet)
	if err != nil {
		return err
	}

	if err := containersRemove(containers, options); err != nil {
		return err
	}

	return nil
}

func werfContainersByFilterSet(filterSet filters.Args) ([]types.Container, error) {
	filterSet.Add("name", image.StageContainerNamePrefix)
	return containersByFilterSet(filterSet)
}

func containersByFilterSet(filterSet filters.Args) ([]types.Container, error) {
	containersOptions := types.ContainerListOptions{}
	containersOptions.All = true
	containersOptions.Quiet = true
	containersOptions.Filters = filterSet

	return docker.Containers(containersOptions)
}

func containersRemove(containers []types.Container, options CommonOptions) error {
	for _, container := range containers {
		if options.DryRun {
			logboek.LogLn(logContainerName(container))
			logboek.LogOptionalLn()
		} else {
			if err := docker.ContainerRemove(container.ID, types.ContainerRemoveOptions{Force: options.RmForce}); err != nil {
				return err
			}
		}
	}

	return nil
}
