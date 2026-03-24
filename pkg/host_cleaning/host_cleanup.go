package host_cleaning

import (
	"context"
	"errors"
	"fmt"

	"github.com/werf/common-go/pkg/graceful"
	"github.com/werf/logboek"
	thresholdpkg "github.com/werf/werf/v2/pkg/cleaning/threshold"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/git_repo/gitdata"
	"github.com/werf/werf/v2/pkg/tmp_manager"
	"github.com/werf/werf/v2/pkg/werf"
	"github.com/werf/werf/v2/pkg/werf/exec"
)

const (
	DefaultAllowedBackendStorageVolumeUsagePercentage       float64 = 70.0
	DefaultAllowedBackendStorageVolumeUsageMarginPercentage float64 = 5.0
	DefaultAllowedLocalCacheVolumeUsagePercentage           float64 = 70.0
	DefaultAllowedLocalCacheVolumeUsageMarginPercentage     float64 = 5.0
)

type VolumeUsageThresholdType = thresholdpkg.Type

const (
	VolumeUsageThresholdTypePercentage = thresholdpkg.TypePercentage
	VolumeUsageThresholdTypeBytes      = thresholdpkg.TypeBytes
)

type VolumeUsageThreshold = thresholdpkg.Threshold

func NewVolumeUsageThresholdPercentage(value uint64) VolumeUsageThreshold {
	return thresholdpkg.NewPercentage(value)
}

func NewVolumeUsageThresholdBytes(value uint64) VolumeUsageThreshold {
	return thresholdpkg.NewBytes(value)
}

func DefaultAllowedBackendStorageVolumeUsageThreshold() VolumeUsageThreshold {
	return NewVolumeUsageThresholdPercentage(uint64(DefaultAllowedBackendStorageVolumeUsagePercentage))
}

func DefaultAllowedBackendStorageVolumeUsageMarginThreshold() VolumeUsageThreshold {
	return NewVolumeUsageThresholdPercentage(uint64(DefaultAllowedBackendStorageVolumeUsageMarginPercentage))
}

func DefaultAllowedLocalCacheVolumeUsageThreshold() VolumeUsageThreshold {
	return NewVolumeUsageThresholdPercentage(uint64(DefaultAllowedLocalCacheVolumeUsagePercentage))
}

func DefaultAllowedLocalCacheVolumeUsageMarginThreshold() VolumeUsageThreshold {
	return NewVolumeUsageThresholdPercentage(uint64(DefaultAllowedLocalCacheVolumeUsageMarginPercentage))
}

func ParseVolumeUsageThreshold(value string) (VolumeUsageThreshold, error) {
	return thresholdpkg.Parse(value)
}

func volumeUsageThresholdValueOrDefault(optionValue *VolumeUsageThreshold, defaultValue VolumeUsageThreshold) VolumeUsageThreshold {
	if optionValue != nil {
		return *optionValue
	}
	return defaultValue
}

func resolveVolumeUsageThresholds(thresholdOption, marginOption *VolumeUsageThreshold, defaultThreshold, defaultMargin VolumeUsageThreshold, marginExplicit bool, thresholdFlagName, marginFlagName string) (VolumeUsageThreshold, VolumeUsageThreshold, error) {
	return thresholdpkg.Resolve(thresholdOption, marginOption, defaultThreshold, defaultMargin, marginExplicit, thresholdFlagName, marginFlagName)
}

func resolveBackendStorageVolumeUsageThresholds(thresholdOption, marginOption *VolumeUsageThreshold, marginExplicit bool) (VolumeUsageThreshold, VolumeUsageThreshold, error) {
	return resolveVolumeUsageThresholds(thresholdOption, marginOption, DefaultAllowedBackendStorageVolumeUsageThreshold(), DefaultAllowedBackendStorageVolumeUsageMarginThreshold(), marginExplicit, "--allowed-backend-storage-volume-usage", "--allowed-backend-storage-volume-usage-margin")
}

func resolveLocalCacheVolumeUsageThresholds(thresholdOption, marginOption *VolumeUsageThreshold, marginExplicit bool) (VolumeUsageThreshold, VolumeUsageThreshold, error) {
	return resolveVolumeUsageThresholds(thresholdOption, marginOption, DefaultAllowedLocalCacheVolumeUsageThreshold(), DefaultAllowedLocalCacheVolumeUsageMarginThreshold(), marginExplicit, "--allowed-local-cache-volume-usage", "--allowed-local-cache-volume-usage-margin")
}

type AutoHostCleanupOptions struct {
	HostCleanupOptions

	TmpDir      *string
	HomeDir     *string
	ProjectName *string
}

type HostCleanupOptions struct {
	BackendStoragePath                              *string
	AllowedBackendStorageVolumeUsageThreshold       *VolumeUsageThreshold
	AllowedBackendStorageVolumeUsageMarginThreshold *VolumeUsageThreshold
	AllowedBackendStorageVolumeUsageMarginExplicit  bool
	AllowedLocalCacheVolumeUsageThreshold           *VolumeUsageThreshold
	AllowedLocalCacheVolumeUsageMarginThreshold     *VolumeUsageThreshold
	AllowedLocalCacheVolumeUsageMarginExplicit      bool

	DryRun bool
	Force  bool
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

	if options.AllowedBackendStorageVolumeUsageThreshold != nil {
		args = append(args, "--allowed-backend-storage-volume-usage", options.AllowedBackendStorageVolumeUsageThreshold.FormatCLIValue())
	}
	if options.AllowedBackendStorageVolumeUsageMarginThreshold != nil {
		args = append(args, "--allowed-backend-storage-volume-usage-margin", options.AllowedBackendStorageVolumeUsageMarginThreshold.FormatCLIValue())
	}
	if options.AllowedLocalCacheVolumeUsageThreshold != nil {
		args = append(args, "--allowed-local-cache-volume-usage", options.AllowedLocalCacheVolumeUsageThreshold.FormatCLIValue())
	}
	if options.AllowedLocalCacheVolumeUsageMarginThreshold != nil {
		args = append(args, "--allowed-local-cache-volume-usage-margin", options.AllowedLocalCacheVolumeUsageMarginThreshold.FormatCLIValue())
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

	allowedLocalCacheVolumeUsageThreshold, allowedLocalCacheVolumeUsageMarginThreshold, err := resolveLocalCacheVolumeUsageThresholds(options.AllowedLocalCacheVolumeUsageThreshold, options.AllowedLocalCacheVolumeUsageMarginThreshold, options.AllowedLocalCacheVolumeUsageMarginExplicit)
	if err != nil {
		return err
	}

	if err := logboek.Context(ctx).Default().LogProcess("Running GC for git data").DoError(func() error {
		if err := gitdata.RunGC(ctx, gitdata.RunGCOptions{
			AllowedLocalCacheVolumeUsageThreshold:       allowedLocalCacheVolumeUsageThreshold,
			AllowedLocalCacheVolumeUsageMarginThreshold: allowedLocalCacheVolumeUsageMarginThreshold,
			DryRun: options.DryRun,
		}); err != nil {
			return fmt.Errorf("git repo GC failed: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}

	allowedBackendStorageVolumeUsageThreshold, allowedBackendStorageVolumeUsageMarginThreshold, err := resolveBackendStorageVolumeUsageThresholds(options.AllowedBackendStorageVolumeUsageThreshold, options.AllowedBackendStorageVolumeUsageMarginThreshold, options.AllowedBackendStorageVolumeUsageMarginExplicit)
	if err != nil {
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
		err := cleaner.RunGC(ctx, RunGCOptions{
			AllowedStorageVolumeUsageThreshold:       allowedBackendStorageVolumeUsageThreshold,
			AllowedStorageVolumeUsageMarginThreshold: allowedBackendStorageVolumeUsageMarginThreshold,
			StoragePath:                              *options.BackendStoragePath,
			Force:                                    options.Force,
			DryRun:                                   options.DryRun,
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

	allowedLocalCacheVolumeUsageThreshold := volumeUsageThresholdValueOrDefault(options.AllowedLocalCacheVolumeUsageThreshold, DefaultAllowedLocalCacheVolumeUsageThreshold())

	shouldRun, err = gitdata.ShouldRunAutoGC(ctx, allowedLocalCacheVolumeUsageThreshold)
	if err != nil {
		return false, fmt.Errorf("failed to check git repo GC: %w", err)
	}
	if shouldRun {
		return true, nil
	}

	allowedBackendStorageVolumeUsageThreshold := volumeUsageThresholdValueOrDefault(options.AllowedBackendStorageVolumeUsageThreshold, DefaultAllowedBackendStorageVolumeUsageThreshold())

	cleaner, err := NewLocalBackendCleaner(backend, werf.HostLocker().Locker())
	if errors.Is(err, ErrUnsupportedContainerBackend) {
		// if cleaner not implemented, skip cleaning
		return false, nil
	} else if err != nil {
		return false, err
	}

	shouldRun, err = cleaner.ShouldRunAutoGC(ctx, RunAutoGCOptions{
		AllowedStorageVolumeUsageThreshold: allowedBackendStorageVolumeUsageThreshold,
		StoragePath:                        *options.BackendStoragePath,
	})
	if err != nil {
		return false, fmt.Errorf("failed to check local %s backend GC: %w", cleaner.BackendName(), err)
	}
	if shouldRun {
		return true, nil
	}

	return false, nil
}
