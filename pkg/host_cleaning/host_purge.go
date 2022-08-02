package host_cleaning

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/docker/docker/api/types/filters"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/container_backend"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/stapel"
	"github.com/werf/werf/pkg/tmp_manager"
	"github.com/werf/werf/pkg/werf"
)

type HostPurgeOptions struct {
	DryRun                        bool
	RmContainersThatUseWerfImages bool
}

func HostPurge(ctx context.Context, containerBackend container_backend.ContainerBackend, options HostPurgeOptions) error {
	commonOptions := CommonOptions{
		RmiForce:                      true,
		RmForce:                       true,
		RmContainersThatUseWerfImages: options.RmContainersThatUseWerfImages,
		DryRun:                        options.DryRun,
	}

	if err := logboek.Context(ctx).LogProcess("Running werf docker containers purge").DoError(func() error {
		if err := werfContainersFlushByFilterSet(ctx, filters.NewArgs(), commonOptions); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	if err := logboek.Context(ctx).LogProcess("Running werf docker images purge").DoError(func() error {
		filterSet := filters.NewArgs()
		filterSet.Add("label", image.WerfLabel)

		if err := werfImagesFlushByFilterSet(ctx, filterSet, commonOptions); err != nil {
			return err
		}

		localManagedImageRecordImageNameFormat := "werf-managed-images/%s" // legacy
		filterSet = filters.NewArgs()
		filterSet.Add("reference", fmt.Sprintf(localManagedImageRecordImageNameFormat, "*"))

		if err := werfImagesFlushByFilterSet(ctx, filterSet, commonOptions); err != nil {
			return err
		}

		localMetadataImageRecordImageNameFormat := "werf-images-metadata-by-commit/%s" // legacy
		filterSet = filters.NewArgs()
		filterSet.Add("reference", fmt.Sprintf(localMetadataImageRecordImageNameFormat, "*"))

		if err := werfImagesFlushByFilterSet(ctx, filterSet, commonOptions); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	if err := tmp_manager.Purge(ctx, commonOptions.DryRun, containerBackend); err != nil {
		return fmt.Errorf("tmp files purge failed: %w", err)
	}

	if err := logboek.Context(ctx).LogProcess("Running werf home data purge").DoError(func() error {
		return purgeHomeWerfFiles(ctx, commonOptions.DryRun, containerBackend)
	}); err != nil {
		return err
	}

	if err := logboek.Context(ctx).LogProcess("Deleting stapel").DoError(func() error {
		return deleteStapel(ctx, commonOptions.DryRun)
	}); err != nil {
		return fmt.Errorf("stapel delete failed: %w", err)
	}

	return nil
}

func deleteStapel(ctx context.Context, dryRun bool) error {
	if dryRun {
		return nil
	}

	if err := stapel.Purge(ctx); err != nil {
		return err
	}

	return nil
}

func purgeHomeWerfFiles(ctx context.Context, dryRun bool, containerBackend container_backend.ContainerBackend) error {
	pathsToRemove := []string{werf.GetServiceDir(), werf.GetLocalCacheDir(), werf.GetSharedContextDir()}

	for _, path := range pathsToRemove {
		logboek.Context(ctx).LogLn(path)
	}

	if dryRun {
		return nil
	}

	if runtime.GOOS == "windows" {
		for _, path := range pathsToRemove {
			if err := os.RemoveAll(path); err != nil {
				return err
			}
		}

		return nil
	} else {
		return containerBackend.RemoveHostDirs(ctx, werf.GetHomeDir(), pathsToRemove)
	}
}
