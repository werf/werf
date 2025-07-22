package host_cleaning

import (
	"context"
	"fmt"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/tmp_manager"
)

type HostPurgeOptions struct {
	DryRun                        bool
	RmContainersThatUseWerfImages bool
}

func HostPurge(ctx context.Context, backend container_backend.ContainerBackend, options HostPurgeOptions) error {
	purger := newLocalPurger(backend)

	commonOptions := CommonOptions{
		RmiForce:                      true,
		RmForce:                       true,
		RmContainersThatUseWerfImages: options.RmContainersThatUseWerfImages,
		DryRun:                        options.DryRun,
	}

	if err := logboek.Context(ctx).LogProcess("Running werf docker containers purge").DoError(func() error {
		return purger.FlushContainers(ctx, commonOptions)
	}); err != nil {
		return err
	}

	if err := logboek.Context(ctx).LogProcess("Running werf docker images purge").DoError(func() error {
		return purger.FlushImages(ctx, commonOptions)
	}); err != nil {
		return err
	}

	if err := tmp_manager.Purge(ctx, commonOptions.DryRun); err != nil {
		return fmt.Errorf("tmp files purge failed: %w", err)
	}

	if err := logboek.Context(ctx).LogProcess("Running werf home data purge").DoError(func() error {
		return purger.PurgeWerfHomeFiles(ctx, commonOptions)
	}); err != nil {
		return err
	}

	if err := logboek.Context(ctx).LogProcess("Deleting stapel").DoError(func() error {
		return purger.PurgeStapelFiles(ctx, commonOptions)
	}); err != nil {
		return fmt.Errorf("stapel delete failed: %w", err)
	}

	return nil
}
