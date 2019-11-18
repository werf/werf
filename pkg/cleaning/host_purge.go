package cleaning

import (
	"fmt"

	"github.com/docker/docker/api/types/filters"

	"github.com/flant/logboek"

	"github.com/flant/werf/pkg/stapel"
	"github.com/flant/werf/pkg/tmp_manager"
	"github.com/flant/werf/pkg/util"
	"github.com/flant/werf/pkg/werf"
)

type HostPurgeOptions struct {
	DryRun                        bool
	RmContainersThatUseWerfImages bool
}

func HostPurge(options HostPurgeOptions) error {
	commonOptions := CommonOptions{
		RmiForce:                      true,
		RmForce:                       true,
		RmContainersThatUseWerfImages: options.RmContainersThatUseWerfImages,
		DryRun:                        options.DryRun,
	}

	if err := logboek.LogProcess("Running werf docker containers purge", logboek.LogProcessOptions{}, func() error {
		if err := werfContainersFlushByFilterSet(filters.NewArgs(), commonOptions); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	if err := logboek.LogProcess("Running werf docker images purge", logboek.LogProcessOptions{}, func() error {
		if err := werfImagesFlushByFilterSet(filters.NewArgs(), commonOptions); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	if err := tmp_manager.Purge(commonOptions.DryRun); err != nil {
		return fmt.Errorf("tmp files purge failed: %s", err)
	}

	if err := logboek.LogProcess("Running werf home data purge", logboek.LogProcessOptions{}, func() error {
		return purgeHomeWerfFiles(commonOptions.DryRun)
	}); err != nil {
		return err
	}

	if err := logboek.LogProcess("Deleting stapel", logboek.LogProcessOptions{}, func() error {
		return deleteStapel(commonOptions.DryRun)
	}); err != nil {
		return fmt.Errorf("stapel delete failed: %s", err)
	}

	return nil
}

func deleteStapel(dryRun bool) error {
	if dryRun {
		return nil
	}

	if err := stapel.Purge(); err != nil {
		return err
	}

	return nil
}

func purgeHomeWerfFiles(dryRun bool) error {
	pathsToRemove := []string{werf.GetServiceDir(), werf.GetLocalCacheDir(), werf.GetSharedContextDir()}

	for _, path := range pathsToRemove {
		logboek.LogLn(path)
	}

	if dryRun {
		return nil
	}

	return util.RemoveHostDirsWithLinuxContainer(werf.GetHomeDir(), pathsToRemove)
}
