package host_cleaning

import (
	"context"
	"fmt"
	"strings"

	chart "github.com/werf/common-go/pkg/lock"
	"github.com/werf/lockgate"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/image"
)

func werfContainersByContainersOptions(ctx context.Context, backend container_backend.ContainerBackend, containersOptions container_backend.ContainersOptions) (image.ContainerList, error) {
	containersOptions.Filters = append(containersOptions.Filters,
		image.ContainerFilter{Name: image.StageContainerNamePrefix})
	return backend.Containers(ctx, containersOptions)
}

func containersRemove(ctx context.Context, backend container_backend.ContainerBackend, containers image.ContainerList, options CommonOptions) error {
	for _, container := range containers {
		if options.DryRun {
			logboek.Context(ctx).LogLn(logContainerName(container))
			logboek.Context(ctx).LogOptionalLn()
		} else {
			err := backend.Rm(ctx, container.ID, container_backend.RmOpts{
				Force: options.RmForce,
			})
			if err != nil {
				return fmt.Errorf("container backend rm: %w", err)
			}
		}
	}

	return nil
}

func werfContainerName(container image.Container) string {
	var containerName string
	for _, name := range container.Names {
		if strings.HasPrefix(name, fmt.Sprintf("/%s", image.StageContainerNamePrefix)) {
			containerName = strings.TrimPrefix(name, "/")
			break
		}
	}
	return containerName
}

func doWithContainerLock(ctx context.Context, container image.Container, fn func() error) error {
	containerName := werfContainerName(container)
	containerLockName := container_backend.ContainerLockName(containerName)

	isLocked, lock, err := chart.AcquireHostLock(ctx, containerLockName, lockgate.AcquireOptions{NonBlocking: true})
	if err != nil {
		return fmt.Errorf("failed to lock %s for container %s: %w", containerLockName, logContainerName(container), err)
	}

	if err != nil {
		return fmt.Errorf("failed to lock %s for container %s: %w", containerLockName, logContainerName(container), err)
	}

	if !isLocked {
		logboek.Context(ctx).LogFDetails("Ignore container %s used by another process\n", logContainerName(container))
		return nil
	}

	defer chart.ReleaseHostLock(lock)

	return fn()
}

func buildContainersOptions(filters ...image.ContainerFilter) container_backend.ContainersOptions {
	opts := container_backend.ContainersOptions{}
	opts.Filters = filters
	return opts
}
