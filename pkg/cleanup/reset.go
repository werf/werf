package cleanup

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types/filters"

	"github.com/flant/werf/pkg/werf"
)

func ResetAll(options CommonOptions) error {
	if err := werfContainersFlushByFilterSet(filters.NewArgs(), options); err != nil {
		return err
	}

	if err := werfImagesFlushByFilterSet(filters.NewArgs(), options); err != nil {
		return err
	}

	if err := deleteWerfFiles(options); err != nil {
		return err
	}

	if err := RemoveLostTmpWerfFiles(); err != nil {
		return err
	}

	return nil
}

func deleteWerfFiles(options CommonOptions) error {
	var directoryPathToDelete []string
	for _, directory := range []string{"bin", "builds", "git", "worktree", "tmp"} {
		directoryPath := filepath.Join(werf.GetHomeDir(), directory)

		if _, err := os.Stat(directoryPath); !os.IsNotExist(err) {
			directoryPathToDelete = append(directoryPathToDelete, directoryPath)
		}
	}

	if len(directoryPathToDelete) != 0 {
		fmt.Println("reset werf cache")
		for _, directoryPath := range directoryPathToDelete {
			if options.DryRun {
				fmt.Println(directoryPath)
			} else {
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
	if err := werfDimgstagesFlushByCacheVersion(filters.NewArgs(), options); err != nil {
		return err
	}

	return nil
}
