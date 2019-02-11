package cleanup

import (
	"fmt"
	"os"

	"github.com/docker/docker/api/types/filters"

	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/logger"
	"github.com/flant/werf/pkg/tmp_manager"
	"github.com/flant/werf/pkg/werf"
)

func HostCleanup(options CommonOptions) error {
	// FIXME: remove only unused garbage containers and images

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

func HostPurge(options CommonOptions) error {
	if err := werfContainersFlushByFilterSet(filters.NewArgs(), options); err != nil {
		return err
	}

	if err := werfImagesFlushByFilterSet(filters.NewArgs(), options); err != nil {
		return err
	}

	if err := tmp_manager.Purge(); err != nil {
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
		fmt.Fprintln(logger.GetOutStream(), "purge werf host data")
		for _, directoryPath := range directoryPathToDelete {
			fmt.Fprintln(logger.GetOutStream(), directoryPath)
			if !options.DryRun {
				err := os.RemoveAll(directoryPath)
				if err != nil {
					return err
				}
			}
		}
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
