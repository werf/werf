package cleanup

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/docker/docker/api/types/filters"
	"github.com/flant/dapp/pkg/dapp"
)

type ResetOptions struct {
	Mode          ResetModeOptions `json:"mode"`
	CacheVersion  string           `json:"cache_version"`
	CommonOptions CommonOptions    `json:"common_options"`
}

type ResetModeOptions struct {
	All          bool `json:"all"`
	DevModeCache bool `json:"dev_mode_cache"`
	CacheVersion bool `json:"cache_version"`
}

func Reset(options ResetOptions) error {
	if options.Mode.All {
		return resetAll(options)
	} else if options.Mode.DevModeCache {
		return resetDevModeCache(options)
	} else if options.Mode.CacheVersion {
		return resetCacheVersion(options)
	} else {
		return fmt.Errorf("expected command option '--improper-dev-mode-cache', '--improper-cache-version-stages' or '--all'") //	TODO
	}

	return nil
}

func resetAll(options ResetOptions) error {
	if err := dappContainersFlushByFilterSet(filters.NewArgs(), options.CommonOptions); err != nil {
		return err
	}

	if err := dappImagesFlushByFilterSet(filters.NewArgs(), options.CommonOptions); err != nil {
		return err
	}

	if err := deleteDappFiles(options.CommonOptions); err != nil {
		return err
	}

	return nil
}

func deleteDappFiles(options CommonOptions) error {
	var directoryPathToDelete []string
	for _, directory := range []string{"bin", "builds", "git"} {
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

func userHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	} else if runtime.GOOS == "linux" {
		home := os.Getenv("XDG_CONFIG_HOME")
		if home != "" {
			return home
		}
	}
	return os.Getenv("HOME")
}

func resetDevModeCache(options ResetOptions) error {
	filterSet := filters.NewArgs()
	filterSet.Add("label", "dapp-dev-mode")
	if err := dappContainersFlushByFilterSet(filterSet, options.CommonOptions); err != nil {
		return err
	}

	filterSet = filters.NewArgs()
	filterSet.Add("label", "dapp-dev-mode")
	if err := dappImagesFlushByFilterSet(filterSet, options.CommonOptions); err != nil {
		return err
	}

	return nil
}

func resetCacheVersion(options ResetOptions) error {
	if err := dappDimgstagesFlushByCacheVersion(filters.NewArgs(), options.CacheVersion, options.CommonOptions); err != nil {
		return err
	}

	return nil
}
