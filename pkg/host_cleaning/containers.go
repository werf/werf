package host_cleaning

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"

	docker_container "github.com/docker/docker/api/types/container"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/docker"
	"github.com/werf/werf/v2/pkg/image"
)

func werfContainersFlushByFilterSet(ctx context.Context, filterSet filters.Args, options CommonOptions) error {
	containers, err := werfContainersByFilterSet(ctx, filterSet)
	if err != nil {
		return err
	}

	if err := containersRemove(ctx, containers, options); err != nil {
		return err
	}

	return nil
}

func werfContainersByFilterSet(ctx context.Context, filterSet filters.Args) ([]types.Container, error) {
	filterSet.Add("name", image.StageContainerNamePrefix)
	return containersByFilterSet(ctx, filterSet)
}

func containersByFilterSet(ctx context.Context, filterSet filters.Args) ([]types.Container, error) {
	containersOptions := docker_container.ListOptions{}
	containersOptions.All = true
	containersOptions.Filters = filterSet

	return docker.Containers(ctx, containersOptions)
}

func containersRemove(ctx context.Context, containers []types.Container, options CommonOptions) error {
	for _, container := range containers {
		if options.DryRun {
			logboek.Context(ctx).LogLn(logContainerName(container))
			logboek.Context(ctx).LogOptionalLn()
		} else {
			if err := docker.ContainerRemove(ctx, container.ID, docker_container.RemoveOptions{Force: options.RmForce}); err != nil {
				return err
			}
		}
	}

	return nil
}
