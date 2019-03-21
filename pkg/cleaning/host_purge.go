package cleaning

import (
	"fmt"

	"github.com/docker/docker/api/types/filters"

	"github.com/flant/logboek"
	"github.com/flant/werf/pkg/stapel"

	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/tmp_manager"
	"github.com/flant/werf/pkg/werf"
)

func HostPurge(options CommonOptions) error {
	options.SkipUsedImages = false
	options.RmiForce = true
	options.RmForce = true

	if err := logboek.LogSecondaryProcess("Running werf docker containers purge", logboek.LogProcessOptions{}, func() error {
		if err := werfContainersFlushByFilterSet(filters.NewArgs(), options); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	if err := logboek.LogSecondaryProcess("Running werf docker images purge", logboek.LogProcessOptions{}, func() error {
		if err := werfImagesFlushByFilterSet(filters.NewArgs(), options); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	if err := tmp_manager.Purge(options.DryRun); err != nil {
		return fmt.Errorf("tmp files purge failed: %s", err)
	}

	if err := logboek.LogSecondaryProcess("Running werf home data purge", logboek.LogProcessOptions{}, func() error {
		return purgeHomeWerfFiles(options.DryRun)
	}); err != nil {
		return err
	}

	if err := logboek.LogSecondaryProcess("Deleting stapel", logboek.LogProcessOptions{}, func() error {
		return deleteStapel(options.DryRun)
	}); err != nil {
		return fmt.Errorf("stapel delete failed: %s", err)
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

	args := []string{
		"--rm",
		"--volume", fmt.Sprintf("%s:%s", werf.GetHomeDir(), werf.GetHomeDir()),
		stapel.ImageName(),
		stapel.RmBinPath(), "-rf",
	}

	args = append(args, pathsToRemove...)

	return docker.CliRun(args...)
}
