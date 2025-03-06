package host_cleaning

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/stapel"
	"github.com/werf/werf/v2/pkg/werf"
)

type localPurger struct {
	backend container_backend.ContainerBackend
}

func newLocalPurger(backend container_backend.ContainerBackend) *localPurger {
	return &localPurger{
		backend: backend,
	}
}

func (purger *localPurger) FlushContainers(ctx context.Context, options CommonOptions) error {
	containers, err := werfContainersByContainersOptions(ctx, purger.backend, buildContainersOptions())
	if err != nil {
		return err
	}

	if err = containersRemove(ctx, purger.backend, containers, options); err != nil {
		return err
	}

	return nil
}

func (purger *localPurger) FlushImages(ctx context.Context, options CommonOptions) error {
	imagesOptions := buildImagesOptions(
		util.NewPair("label", image.WerfLabel),
	)
	if err := werfImagesFlushByFilterSet(ctx, purger.backend, imagesOptions, options); err != nil {
		return err
	}

	imagesOptions = buildImagesOptions(
		util.NewPair("reference", fmt.Sprintf("werf-managed-images/%s", "*")), // legacy
	)
	if err := werfImagesFlushByFilterSet(ctx, purger.backend, imagesOptions, options); err != nil {
		return err
	}

	imagesOptions = buildImagesOptions(
		util.NewPair("reference", fmt.Sprintf("werf-images-metadata-by-commit/%s", "*")), // legacy
	)
	if err := werfImagesFlushByFilterSet(ctx, purger.backend, imagesOptions, options); err != nil {
		return err
	}

	return nil
}

func (purger *localPurger) PurgeWerfHomeFiles(ctx context.Context, options CommonOptions) error {
	pathsToRemove := []string{werf.GetServiceDir(), werf.GetLocalCacheDir(), werf.GetSharedContextDir()}

	for _, path := range pathsToRemove {
		logboek.Context(ctx).LogLn(path)
	}

	if options.DryRun {
		return nil
	}

	if runtime.GOOS == "windows" {
		for _, path := range pathsToRemove {
			if err := os.RemoveAll(path); err != nil {
				return fmt.Errorf("os remove_all: %w", err)
			}
		}

		return nil
	}

	err := purger.backend.RemoveHostDirs(ctx, werf.GetHomeDir(), pathsToRemove)
	if err != nil {
		return fmt.Errorf("container_backend remove host dirs: %w", err)
	}

	return nil
}

func (purger *localPurger) PurgeStapelFiles(ctx context.Context, options CommonOptions) error {
	if options.DryRun {
		return nil
	}

	err := stapel.Purge(ctx)
	if err != nil {
		return fmt.Errorf("stapel purge: %w", err)
	}

	return nil
}
