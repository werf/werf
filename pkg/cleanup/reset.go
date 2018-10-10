package cleanup

import (
	"fmt"

	"github.com/docker/docker/api/types/filters"
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

	return nil
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
