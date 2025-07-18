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
	"github.com/werf/werf/v2/pkg/werf/exec"
)

const (
	DefaultAllowedBackendStorageVolumeUsagePercentage       float64 = 70.0
	DefaultAllowedBackendStorageVolumeUsageMarginPercentage float64 = 5.0
	DefaultAllowedLocalCacheVolumeUsagePercentage           float64 = 70.0
	DefaultAllowedLocalCacheVolumeUsageMarginPercentage     float64 = 5.0
)

type AutoHostCleanupOptions struct {
	HostCleanupOptions

	TmpDir      *string
	HomeDir     *string
	ProjectName *string
}

type HostCleanupOptions struct {
	BackendStoragePath                               *string
	AllowedBackendStorageVolumeUsagePercentage       *uint
	AllowedBackendStorageVolumeUsageMarginPercentage *uint
	AllowedLocalCacheVolumeUsagePercentage           *uint
	AllowedLocalCacheVolumeUsageMarginPercentage     *uint

	DryRun bool
	Force  bool
}

func getOptionValueOrDefault(optionValue *uint, defaultValue float64) float64 {
	var res float64
	if optionValue != nil {
		res = float64(*optionValue)
	} else {
		res = defaultValue
	}
	return res
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

	if options.AllowedBackendStorageVolumeUsagePercentage != nil {
		args = append(args, "--allowed-backend-storage-volume-usage", fmt.Sprintf("%d", *options.AllowedBackendStorageVolumeUsagePercentage))
	}
	if options.AllowedBackendStorageVolumeUsageMarginPercentage != nil {
		args = append(args, "--allowed-backend-storage-volume-usage-margin", fmt.Sprintf("%d", *options.AllowedBackendStorageVolumeUsageMarginPercentage))
	}
	if options.AllowedLocalCacheVolumeUsagePercentage != nil {
		args = append(args, "--allowed-local-cache-volume-usage", fmt.Sprintf("%d", *options.AllowedLocalCacheVolumeUsagePercentage))
	}
	if options.AllowedLocalCacheVolumeUsageMarginPercentage != nil {
		args = append(args, "--allowed-local-cache-volume-usage-margin", fmt.Sprintf("%d", *options.AllowedLocalCacheVolumeUsageMarginPercentage))
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

	allowedLocalCacheVolumeUsagePercentage := getOptionValueOrDefault(options.AllowedLocalCacheVolumeUsagePercentage, DefaultAllowedLocalCacheVolumeUsagePercentage)
	allowedLocalCacheVolumeUsageMarginPercentage := getOptionValueOrDefault(options.AllowedLocalCacheVolumeUsageMarginPercentage, DefaultAllowedLocalCacheVolumeUsageMarginPercentage)

	if err := logboek.Context(ctx).Default().LogProcess("Running GC for git data").DoError(func() error {
		if err := gitdata.RunGC(ctx, allowedLocalCacheVolumeUsagePercentage, allowedLocalCacheVolumeUsageMarginPercentage); err != nil {
			return fmt.Errorf("git repo GC failed: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}

	allowedBackendStorageVolumeUsagePercentage := getOptionValueOrDefault(options.AllowedBackendStorageVolumeUsagePercentage, DefaultAllowedBackendStorageVolumeUsagePercentage)
	allowedBackendStorageVolumeUsageMarginPercentage := getOptionValueOrDefault(options.AllowedBackendStorageVolumeUsageMarginPercentage, DefaultAllowedBackendStorageVolumeUsageMarginPercentage)

	cleaner, err := NewLocalBackendCleaner(backend)
	if errors.Is(err, ErrUnsupportedContainerBackend) {
		// if cleaner not implemented, skip cleaning
		return nil
	} else if err != nil {
		return err
	}

	return logboek.Context(ctx).Default().LogProcess("Running GC for local %s backend", cleaner.BackendName()).DoError(func() error {
		err := cleaner.RunGC(ctx, RunGCOptions{
			AllowedStorageVolumeUsagePercentage:       allowedBackendStorageVolumeUsagePercentage,
			AllowedStorageVolumeUsageMarginPercentage: allowedBackendStorageVolumeUsageMarginPercentage,
			StoragePath: *options.BackendStoragePath,
			Force:       options.Force,
			DryRun:      options.DryRun,
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

	allowedLocalCacheVolumeUsagePercentage := getOptionValueOrDefault(options.AllowedLocalCacheVolumeUsagePercentage, DefaultAllowedLocalCacheVolumeUsagePercentage)

	shouldRun, err = gitdata.ShouldRunAutoGC(ctx, allowedLocalCacheVolumeUsagePercentage)
	if err != nil {
		return false, fmt.Errorf("failed to check git repo GC: %w", err)
	}
	if shouldRun {
		return true, nil
	}

	allowedBackendStorageVolumeUsagePercentage := getOptionValueOrDefault(options.AllowedBackendStorageVolumeUsagePercentage, DefaultAllowedBackendStorageVolumeUsagePercentage)

	cleaner, err := NewLocalBackendCleaner(backend)
	if errors.Is(err, ErrUnsupportedContainerBackend) {
		// if cleaner not implemented, skip cleaning
		return false, nil
	} else if err != nil {
		return false, err
	}

	shouldRun, err = cleaner.ShouldRunAutoGC(ctx, RunAutoGCOptions{
		AllowedStorageVolumeUsagePercentage: allowedBackendStorageVolumeUsagePercentage,
		StoragePath:                         *options.BackendStoragePath,
	})
	if err != nil {
		return false, fmt.Errorf("failed to check local docker server host cleaner GC: %w", err)
	}
	if shouldRun {
		return true, nil
	}

	return false, nil
}
