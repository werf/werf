package host_cleaning

import (
	"context"
	"errors"
	"fmt"

	"github.com/werf/common-go/pkg/graceful"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/git_repo/gitdata"
	"github.com/werf/werf/v2/pkg/tmp_manager"
	"github.com/werf/werf/v2/cmd/werf/common/units"
	"github.com/werf/werf/v2/pkg/volumeutils"
	"github.com/werf/werf/v2/pkg/werf"
	"github.com/werf/werf/v2/pkg/werf/exec"
)

const (
	DefaultAllowedBackendStorageVolumeUsagePercentage       uint64 = 70
	DefaultAllowedBackendStorageVolumeUsageMarginPercentage uint64 = 5
	DefaultAllowedLocalCacheVolumeUsagePercentage           uint64 = 70
	DefaultAllowedLocalCacheVolumeUsageMarginPercentage     uint64 = 5
)

type AutoHostCleanupOptions struct {
	HostCleanupOptions

	TmpDir      *string
	HomeDir     *string
	ProjectName *string
}

type HostCleanupOptions struct {
	BackendStoragePath                     *string
	AllowedBackendStorageVolumeUsage       *units.UnitValue
	AllowedBackendStorageVolumeUsageMargin *units.UnitValue
	AllowedLocalCacheVolumeUsage           *units.UnitValue
	AllowedLocalCacheVolumeUsageMargin     *units.UnitValue

	DryRun bool
	Force  bool
}

func getRequirementInBytes(val *units.UnitValue, defaultPercent uint64, totalBytes uint64) uint64 {
	if val != nil {
		return val.ToBytes(totalBytes)
	}
	return (totalBytes * defaultPercent) / 100
}

func RunAutoHostCleanup(ctx context.Context, backend container_backend.ContainerBackend, options AutoHostCleanupOptions) error {
	if shouldRun, err := shouldRunAutoHostCleanup(ctx, backend, options); err != nil {
		logboek.Context(ctx).Warn().LogF("WARNING: unable to check if auto host cleanup should be run: %s\n", err)
		return nil
	} else if !shouldRun {
		return nil
	}

	logboek.Context(ctx).Debug().LogF("RunAutoHostCleanup forking ...\n")

	var args []string

	args = append(args,
		"host", "cleanup",
		fmt.Sprintf("--dry-run=%v", options.DryRun),
		fmt.Sprintf("--force=%v", options.Force),
	)

	if options.AllowedBackendStorageVolumeUsage != nil {
		args = append(args, "--allowed-backend-storage-volume-usage", options.AllowedBackendStorageVolumeUsage.String())
	}
	if options.AllowedBackendStorageVolumeUsageMargin != nil {
		args = append(args, "--allowed-backend-storage-volume-usage-margin", options.AllowedBackendStorageVolumeUsageMargin.String())
	}
	if options.AllowedLocalCacheVolumeUsage != nil {
		args = append(args, "--allowed-local-cache-volume-usage", options.AllowedLocalCacheVolumeUsage.String())
	}
	if options.AllowedLocalCacheVolumeUsageMargin != nil {
		args = append(args, "--allowed-local-cache-volume-usage-margin", options.AllowedLocalCacheVolumeUsageMargin.String())
	}
	if options.BackendStoragePath != nil && *options.BackendStoragePath != "" {
		args = append(args, "--backend-storage-path", *options.BackendStoragePath)
	}

	// We should pass tmpDir and homeDir via environment variables
	// because background.TryLock() uses them before parsing cli options doing werf.Init().
	var envs []string

	if options.TmpDir != nil && *options.TmpDir != "" {
		envs = append(envs, fmt.Sprintf("WERF_TMP_DIR=%v", options.TmpDir))
	}
	if options.HomeDir != nil && *options.HomeDir != "" {
		envs = append(envs, fmt.Sprintf("WERF_HOME=%v", options.HomeDir))
	}

	return exec.Detach(ctx, args, envs)
}

func RunHostCleanup(ctx context.Context, backend container_backend.ContainerBackend, options HostCleanupOptions) error {
	if err := logboek.Context(ctx).LogProcess("Running GC for tmp data").DoError(func() error {
		if err := tmp_manager.RunGC(ctx, options.DryRun); errors.Is(err, tmp_manager.ErrPathRemoval) {
			return nil
		} else if err != nil {
			return fmt.Errorf("tmp files GC failed: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}

	vuLocalCache, err := volumeutils.GetVolumeUsageByPath(ctx, werf.GetLocalCacheDir())
	if err != nil {
		return fmt.Errorf("error getting local cache volume usage: %w", err)
	}

	allowedLocalCacheVolumeUsageBytes := getRequirementInBytes(options.AllowedLocalCacheVolumeUsage, DefaultAllowedLocalCacheVolumeUsagePercentage, vuLocalCache.TotalBytes)
	allowedLocalCacheVolumeUsageMarginBytes := getRequirementInBytes(options.AllowedLocalCacheVolumeUsageMargin, DefaultAllowedLocalCacheVolumeUsageMarginPercentage, vuLocalCache.TotalBytes)

	if err := logboek.Context(ctx).Default().LogProcess("Running GC for git data").DoError(func() error {
		if err := gitdata.RunGC(ctx, gitdata.RunGCOptions{
			AllowedLocalCacheVolumeUsageBytes:       allowedLocalCacheVolumeUsageBytes,
			AllowedLocalCacheVolumeUsageMarginBytes: allowedLocalCacheVolumeUsageMarginBytes,
			DryRun:                                  options.DryRun,
		}); err != nil {
			return fmt.Errorf("git repo GC failed: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}

	cleaner, err := NewLocalBackendCleaner(backend, werf.HostLocker().Locker())
	if errors.Is(err, ErrUnsupportedContainerBackend) {
		// if cleaner not implemented, skip cleaning
		return nil
	} else if err != nil {
		return err
	}

	return logboek.Context(ctx).Default().LogProcess("Running GC for local %s backend", cleaner.BackendName()).DoError(func() error {
		backendStoragePath, err := cleaner.backendStoragePath(ctx, *options.BackendStoragePath)
		if err != nil {
			return fmt.Errorf("error getting backend storage path: %w", err)
		}

		vuBackend, err := volumeutils.GetVolumeUsageByPath(ctx, backendStoragePath)
		if err != nil {
			return fmt.Errorf("error getting backend volume usage: %w", err)
		}

		allowedBackendStorageVolumeUsageBytes := getRequirementInBytes(options.AllowedBackendStorageVolumeUsage, DefaultAllowedBackendStorageVolumeUsagePercentage, vuBackend.TotalBytes)
		allowedBackendStorageVolumeUsageMarginBytes := getRequirementInBytes(options.AllowedBackendStorageVolumeUsageMargin, DefaultAllowedBackendStorageVolumeUsageMarginPercentage, vuBackend.TotalBytes)

		err = cleaner.RunGC(ctx, RunGCOptions{
			AllowedStorageVolumeUsageBytes:       allowedBackendStorageVolumeUsageBytes,
			AllowedStorageVolumeUsageMarginBytes: allowedBackendStorageVolumeUsageMarginBytes,
			StoragePath:                          *options.BackendStoragePath,
			Force:                                options.Force,
			DryRun:                               options.DryRun,
		})
		if err != nil {
			return fmt.Errorf("local %s backend GC failed: %w", cleaner.BackendName(), err)
		}
		return nil
	})
}

func shouldRunAutoHostCleanup(ctx context.Context, backend container_backend.ContainerBackend, options AutoHostCleanupOptions) (bool, error) {
	if graceful.IsTerminating(ctx) {
		return false, nil
	}
	// host cleanup is not supported for certain project
	if options.ProjectName != nil && *options.ProjectName != "" {
		return false, nil
	}

	shouldRun, err := tmp_manager.ShouldRunAutoGC()
	if err != nil {
		return false, fmt.Errorf("failed to check tmp manager GC: %w", err)
	}
	if shouldRun {
		return true, nil
	}

	vuLocalCache, err := volumeutils.GetVolumeUsageByPath(ctx, werf.GetLocalCacheDir())
	if err != nil {
		return false, fmt.Errorf("error getting local cache volume usage: %w", err)
	}
	allowedLocalCacheVolumeUsageBytes := getRequirementInBytes(options.AllowedLocalCacheVolumeUsage, DefaultAllowedLocalCacheVolumeUsagePercentage, vuLocalCache.TotalBytes)

	shouldRun, err = gitdata.ShouldRunAutoGC(ctx, allowedLocalCacheVolumeUsageBytes)
	if err != nil {
		return false, fmt.Errorf("failed to check git repo GC: %w", err)
	}
	if shouldRun {
		return true, nil
	}

	cleaner, err := NewLocalBackendCleaner(backend, werf.HostLocker().Locker())
	if errors.Is(err, ErrUnsupportedContainerBackend) {
		// if cleaner not implemented, skip cleaning
		return false, nil
	} else if err != nil {
		return false, err
	}

	backendStoragePath, err := cleaner.backendStoragePath(ctx, *options.BackendStoragePath)
	if err != nil {
		return false, fmt.Errorf("error getting backend storage path: %w", err)
	}

	vuBackend, err := volumeutils.GetVolumeUsageByPath(ctx, backendStoragePath)
	if err != nil {
		return false, fmt.Errorf("error getting backend volume usage: %w", err)
	}
	allowedBackendStorageVolumeUsageBytes := getRequirementInBytes(options.AllowedBackendStorageVolumeUsage, DefaultAllowedBackendStorageVolumeUsagePercentage, vuBackend.TotalBytes)

	shouldRun, err = cleaner.ShouldRunAutoGC(ctx, RunAutoGCOptions{
		AllowedStorageVolumeUsageBytes: allowedBackendStorageVolumeUsageBytes,
		StoragePath:                    *options.BackendStoragePath,
	})
	if err != nil {
		return false, fmt.Errorf("failed to check local %s backend GC: %w", cleaner.BackendName(), err)
	}
	if shouldRun {
		return true, nil
	}

	return false, nil
}
