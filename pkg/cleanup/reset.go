package cleanup

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types/filters"
	"github.com/flant/dapp/pkg/dapp"
)

func ResetAll(options CommonOptions) error {
	if err := dappContainersFlushByFilterSet(filters.NewArgs(), options); err != nil {
		return err
	}

	if err := dappImagesFlushByFilterSet(filters.NewArgs(), options); err != nil {
		return err
	}

	if err := deleteDappFiles(options); err != nil {
		return err
	}

	return nil
}

func deleteDappFiles(options CommonOptions) error {
	var directoryPathToDelete []string
	for _, directory := range []string{"bin", "builds", "git", "worktree"} {
		directoryPath := filepath.Join(dapp.GetHomeDir(), directory)

		if _, err := os.Stat(directoryPath); !os.IsNotExist(err) {
			directoryPathToDelete = append(directoryPathToDelete, directoryPath)
		}
	}

	if len(directoryPathToDelete) != 0 {
		fmt.Println("reset dapp cache")
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
	filterSet.Add("label", "dapp-dev-mode")
	if err := dappContainersFlushByFilterSet(filterSet, options); err != nil {
		return err
	}

	filterSet = filters.NewArgs()
	filterSet.Add("label", "dapp-dev-mode")
	if err := dappImagesFlushByFilterSet(filterSet, options); err != nil {
		return err
	}

	return nil
}

func ResetCacheVersion(options CommonOptions) error {
	if err := dappDimgstagesFlushByCacheVersion(filters.NewArgs(), options); err != nil {
		return err
	}

	return nil
}
