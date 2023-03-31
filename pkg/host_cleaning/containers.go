package host_cleaning

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/image"
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
	containersOptions := types.ContainerListOptions{}
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
			if err := docker.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{Force: options.RmForce}); err != nil {
				return err
			}
		}
	}

	return nil
}
