package host_cleaning

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/werf/lockgate"
	"github.com/werf/werf/pkg/werf"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/tmp_manager"
)

type HostCleanupOptions struct {
	DryRun bool
}

func HostCleanup(ctx context.Context, options HostCleanupOptions) error {
	commonOptions := CommonOptions{
		SkipUsedImages: true,
		RmiForce:       false,
		RmForce:        true,
		DryRun:         options.DryRun,
	}

	return werf.WithHostLock(ctx, "host-cleanup", lockgate.AcquireOptions{Timeout: time.Second * 600}, func() error {
		if err := logboek.Context(ctx).LogProcess("Running cleanup for docker containers created by werf").DoError(func() error {
			return safeContainersCleanup(ctx, commonOptions)
		}); err != nil {
			return err
		}

		if err := logboek.Context(ctx).LogProcess("Running cleanup for dangling docker images created by werf").DoError(func() error {
			return safeDanglingImagesCleanup(ctx, commonOptions)
		}); err != nil {
			return nil
		}

		return werf.WithHostLock(ctx, "gc", lockgate.AcquireOptions{}, func() error {
			if err := tmp_manager.GC(ctx, commonOptions.DryRun); err != nil {
				return fmt.Errorf("tmp files gc failed: %s", err)
			}

			return nil
		})
	})
}

func safeDanglingImagesCleanup(ctx context.Context, options CommonOptions) error {
	images, err := werfImagesByFilterSet(ctx, danglingFilterSet())
	if err != nil {
		return err
	}

	var imagesToRemove []types.ImageSummary

	for _, img := range images {
		if imgName, hasKey := img.Labels[image.WerfDockerImageName]; hasKey {
			imageLockName := container_runtime.ImageLockName(imgName)
			isLocked, lock, err := werf.AcquireHostLock(ctx, imageLockName, lockgate.AcquireOptions{NonBlocking: true})
			if err != nil {
				return fmt.Errorf("failed to lock %s for image %s: %s", imageLockName, imgName, err)
			}

			if !isLocked {
				logboek.Context(ctx).Debug().LogFDetails("Ignore dangling image %s processed by another werf process\n", imgName)
				continue
			}

			werf.ReleaseHostLock(lock) // no need to hold a lock

			imagesToRemove = append(imagesToRemove, img)
		} else {
			imagesToRemove = append(imagesToRemove, img)
		}
	}

	imagesToRemove, err = processUsedImages(ctx, imagesToRemove, options)
	if err != nil {
		return err
	}

	if err := imagesRemove(ctx, imagesToRemove, options); err != nil {
		return err
	}

	return nil
}

func safeContainersCleanup(ctx context.Context, options CommonOptions) error {
	containers, err := werfContainersByFilterSet(ctx, filters.NewArgs())
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
			logboek.Context(ctx).Warn().LogF("Ignore bad container %s\n", container.ID)
			continue
		}

		err := func() error {
			containerLockName := container_runtime.ContainerLockName(containerName)
			isLocked, lock, err := werf.AcquireHostLock(ctx, containerLockName, lockgate.AcquireOptions{NonBlocking: true})
			if err != nil {
				return fmt.Errorf("failed to lock %s for container %s: %s", containerLockName, logContainerName(container), err)
			}

			if !isLocked {
				logboek.Context(ctx).Default().LogFDetails("Ignore container %s used by another process\n", logContainerName(container))
				return nil
			}
			defer werf.ReleaseHostLock(lock)

			if err := containersRemove(ctx, []types.Container{container}, options); err != nil {
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
