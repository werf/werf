package cleanup

import (
	"fmt"
	"os"

	"github.com/docker/docker/api/types/filters"

	"github.com/flant/werf/pkg/logger"
	"github.com/flant/werf/pkg/tmp_manager"
	"github.com/flant/werf/pkg/werf"
)

func HostPurge(options CommonOptions) error {
	err := logger.LogServiceProcess("Running werf docker containers purge", logger.LogProcessOptions{}, func() error {
		if err := werfContainersFlushByFilterSet(filters.NewArgs(), options); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	err = logger.LogServiceProcess("Running werf docker images purge", logger.LogProcessOptions{}, func() error {
		if err := werfImagesFlushByFilterSet(filters.NewArgs(), options); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	if err := tmp_manager.Purge(options.DryRun); err != nil {
		return fmt.Errorf("tmp files purge failed: %s", err)
	}

	if err := deleteWerfFiles(options); err != nil {
		return err
	}

	return nil
}

func deleteWerfFiles(options CommonOptions) error {
	var directoryPathToDelete []string
	for _, directoryPath := range []string{werf.GetServiceDir(), werf.GetLocalCacheDir(), werf.GetSharedContextDir()} {
		if _, err := os.Stat(directoryPath); !os.IsNotExist(err) {
			directoryPathToDelete = append(directoryPathToDelete, directoryPath)
		}
	}

	if len(directoryPathToDelete) != 0 {
		return logger.LogServiceProcess("Running werf host data purge", logger.LogProcessOptions{}, func() error {
			for _, directoryPath := range directoryPathToDelete {
				logger.LogF("Removing %s ...\n", directoryPath)
				if !options.DryRun {
					err := os.RemoveAll(directoryPath)
					if err != nil {
						return err
					}
				}
			}

			return nil
		})
	}

	return nil
}

func ResetDevModeCache(options CommonOptions) error {
	filterSet := filters.NewArgs()
	filterSet.Add("label", "werf-dev-mode")
	if err := werfContainersFlushByFilterSet(filterSet, options); err != nil {
		return err
	}

	filterSet = filters.NewArgs()
	filterSet.Add("label", "werf-dev-mode")
	if err := werfImagesFlushByFilterSet(filterSet, options); err != nil {
		return err
	}

	return nil
}

func ResetCacheVersion(options CommonOptions) error {
	if err := werfImageStagesFlushByCacheVersion(filters.NewArgs(), options); err != nil {
		return err
	}

	return nil
}
