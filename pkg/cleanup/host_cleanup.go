package cleanup

import (
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/logger"
	"github.com/flant/werf/pkg/tmp_manager"
)

func HostCleanup(options CommonOptions) error {
	filterSet := filters.NewArgs()
	filterSet.Add("name", image.StageContainerNamePrefix)
	containers, err := containersByFilterSet(filterSet)
	if err != nil {
		return fmt.Errorf("cannot get stages build containers: %s", err)
	}

	for _, container := range containers {
		var containerName string
		for _, name := range container.Names {
			if strings.HasPrefix(name, fmt.Sprintf("/%s", image.StageContainerNamePrefix)) {
				containerName = strings.TrimPrefix(name, "/")
				break
			}
		}

		if containerName == "" {
			logger.LogErrorF("Ignore bad container %s\n", container.ID)
			continue
		}

		err := func() error {
			containerLockName := image.GetContainerLockName(containerName)

			isLocked, err := lock.TryLock(containerLockName, lock.TryLockOptions{})
			if err != nil {
				return fmt.Errorf("failed to lock %s for container %s: %s", containerLockName, containerName, err)
			}

			if !isLocked {
				logger.LogInfoF("Ignore container %s (%s) used by another process\n", containerName, container.ID)
				return nil
			}
			defer lock.Unlock(containerLockName)

			logger.LogInfoF("Removing container %s (%s)\n", containerName, container.ID)

			if !options.DryRun {
				if err := docker.ContainerRemove(container.ID, types.ContainerRemoveOptions{Force: true}); err != nil {
					return fmt.Errorf("failed to remove container %s (%s) :%s", containerName, container.ID, err)
				}
			}

			return nil
		}()

		if err != nil {
			return err
		}
	}

	// if err := werfContainersFlushByFilterSet(filters.NewArgs(), options); err != nil {
	// 	return err
	// }

	// if err := werfImagesFlushByFilterSet(filters.NewArgs(), options); err != nil {
	// 	return err
	// }

	return lock.WithLock("gc", lock.LockOptions{}, func() error {
		if err := tmp_manager.GC(); err != nil {
			return fmt.Errorf("tmp files gc failed: %s", err)
		}

		return nil
	})
}
