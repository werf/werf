package cleaning

import (
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"

	"github.com/flant/logboek"
	"github.com/flant/shluz"
	"github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/tmp_manager"
)

type HostCleanupOptions struct {
	DryRun bool
}

func HostCleanup(options HostCleanupOptions) error {
	commonOptions := CommonOptions{
		SkipUsedImages: true,
		RmiForce:       false,
		RmForce:        true,
		DryRun:         options.DryRun,
	}

	return shluz.WithLock("host-cleanup", shluz.LockOptions{Timeout: time.Second * 600}, func() error {
		if err := logboek.LogProcess("Running cleanup for docker containers created by werf", logboek.LogProcessOptions{}, func() error {
			return safeContainersCleanup(commonOptions)
		}); err != nil {
			return err
		}

		if err := logboek.LogProcess("Running cleanup for dangling docker images created by werf", logboek.LogProcessOptions{}, func() error {
			return safeDanglingImagesCleanup(commonOptions)
		}); err != nil {
			return nil
		}

		return shluz.WithLock("gc", shluz.LockOptions{}, func() error {
			if err := tmp_manager.GC(commonOptions.DryRun); err != nil {
				return fmt.Errorf("tmp files gc failed: %s", err)
			}

			return nil
		})
	})
}

func safeDanglingImagesCleanup(options CommonOptions) error {
	images, err := werfImagesByFilterSet(danglingFilterSet())
	if err != nil {
		return err
	}

	var imagesToRemove []types.ImageSummary

	for _, img := range images {
		if imgName, hasKey := img.Labels[image.WerfDockerImageName]; hasKey {
			imageLockName := image.ImageLockName(imgName)
			isLocked, err := shluz.TryLock(imageLockName, shluz.TryLockOptions{})
			if err != nil {
				return fmt.Errorf("failed to lock %s for image %s: %s", imageLockName, imgName, err)
			}

			if !isLocked {
				logboek.Debug.LogFDetails("Ignore dangling image %s used by another process\n", imgName)
				continue
			}

			shluz.Unlock(imageLockName) // no need to hold a lock

			imagesToRemove = append(imagesToRemove, img)
		} else {
			imagesToRemove = append(imagesToRemove, img)
		}
	}

	imagesToRemove, err = processUsedImages(imagesToRemove, options)
	if err != nil {
		return err
	}

	if err := imagesRemove(imagesToRemove, options); err != nil {
		return err
	}

	return nil
}

func safeContainersCleanup(options CommonOptions) error {
	containers, err := werfContainersByFilterSet(filters.NewArgs())
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
			logboek.LogWarnF("Ignore bad container %s\n", container.ID)
			continue
		}

		err := func() error {
			containerLockName := image.ContainerLockName(containerName)
			isLocked, err := shluz.TryLock(containerLockName, shluz.TryLockOptions{})
			if err != nil {
				return fmt.Errorf("failed to lock %s for container %s: %s", containerLockName, logContainerName(container), err)
			}

			if !isLocked {
				logboek.Default.LogFDetails("Ignore container %s used by another process\n", logContainerName(container))
				return nil
			}
			defer shluz.Unlock(containerLockName)

			if err := containersRemove([]types.Container{container}, options); err != nil {
				return fmt.Errorf("failed to remove container %s: %s", logContainerName(container), err)
			}

			return nil
		}()

		if err != nil {
			return err
		}
	}

	return nil
}
