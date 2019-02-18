package cleaning

import (
	"fmt"

	"github.com/docker/docker/api/types/filters"

	"github.com/flant/werf/pkg/dappdeps"
	"github.com/flant/werf/pkg/docker"
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

	return logger.LogServiceProcess("Running werf home data purge", logger.LogProcessOptions{}, func() error {
		return purgeHomeWerfFiles(options.DryRun)
	})

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

func purgeHomeWerfFiles(dryRun bool) error {
	pathsToRemove := []string{werf.GetServiceDir(), werf.GetLocalCacheDir(), werf.GetSharedContextDir()}

	for _, path := range pathsToRemove {
		logger.LogF("Removing %s ...\n", path)
	}

	if dryRun {
		return nil
	}

	toolchainContainerName, err := dappdeps.ToolchainContainer()
	if err != nil {
		return err
	}

	args := []string{
		"--rm",
		"--volumes-from", toolchainContainerName,
		"--volume", fmt.Sprintf("%s:%s", werf.GetHomeDir(), werf.GetHomeDir()),
		dappdeps.BaseImageName(),
		dappdeps.RmBinPath(), "-rf",
	}

	args = append(args, pathsToRemove...)

	return docker.CliRun(args...)
}
