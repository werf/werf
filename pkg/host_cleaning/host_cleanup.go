package host_cleaning

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/git_repo/gitdata"
	"github.com/werf/werf/v2/pkg/tmp_manager"
)

const (
	DefaultAllowedDockerStorageVolumeUsagePercentage       float64 = 70.0
	DefaultAllowedDockerStorageVolumeUsageMarginPercentage float64 = 5.0
	DefaultAllowedLocalCacheVolumeUsagePercentage          float64 = 70.0
	DefaultAllowedLocalCacheVolumeUsageMarginPercentage    float64 = 5.0
)

type HostCleanupOptions struct {
	AllowedDockerStorageVolumeUsagePercentage       *uint
	AllowedDockerStorageVolumeUsageMarginPercentage *uint
	AllowedLocalCacheVolumeUsagePercentage          *uint
	AllowedLocalCacheVolumeUsageMarginPercentage    *uint
	DockerServerStoragePath                         *string

	DryRun bool
	Force  bool
}

type AutoHostCleanupOptions struct {
	HostCleanupOptions

	ForceShouldRun bool
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
	if !options.ForceShouldRun {
		shouldRun, err := ShouldRunAutoHostCleanup(ctx, backend, options.HostCleanupOptions)
		if err != nil {
			logboek.Context(ctx).Warn().LogF("WARNING: unable to check if auto host cleanup should be run: %s\n", err)
			return nil
		}
		if !shouldRun {
			return nil
		}
	}

	logboek.Context(ctx).Debug().LogF("RunAutoHostCleanup forking ...\n")

	var args []string

	args = append(args,
		"host", "cleanup",
		fmt.Sprintf("--dry-run=%v", options.DryRun),
		fmt.Sprintf("--force=%v", options.Force),
	)

	if options.AllowedDockerStorageVolumeUsagePercentage != nil {
		args = append(args, "--allowed-docker-storage-volume-usage", fmt.Sprintf("%d", *options.AllowedDockerStorageVolumeUsagePercentage))
	}
	if options.AllowedDockerStorageVolumeUsageMarginPercentage != nil {
		args = append(args, "--allowed-docker-storage-volume-usage-margin", fmt.Sprintf("%d", *options.AllowedDockerStorageVolumeUsageMarginPercentage))
	}
	if options.AllowedLocalCacheVolumeUsagePercentage != nil {
		args = append(args, "--allowed-local-cache-volume-usage", fmt.Sprintf("%d", *options.AllowedLocalCacheVolumeUsagePercentage))
	}
	if options.AllowedLocalCacheVolumeUsageMarginPercentage != nil {
		args = append(args, "--allowed-local-cache-volume-usage-margin", fmt.Sprintf("%d", *options.AllowedLocalCacheVolumeUsageMarginPercentage))
	}
	if options.DockerServerStoragePath != nil && *options.DockerServerStoragePath != "" {
		args = append(args, "--docker-server-storage-path", *options.DockerServerStoragePath)
	}

	executableName := os.Getenv("WERF_ORIGINAL_EXECUTABLE")
	if executableName == "" {
		executableName = os.Args[0]
	}

	cmd := exec.Command(executableName, args...)

	var env []string
	for _, spec := range os.Environ() {
		k := strings.SplitN(spec, "=", 2)[0]
		if k == "WERF_ENABLE_PROCESS_EXTERMINATOR" {
			continue
		}

		env = append(env, spec)
	}
	env = append(env, "_WERF_BACKGROUND_MODE_ENABLED=1")

	cmd.Env = env

	if err := cmd.Start(); err != nil {
		logboek.Context(ctx).Warn().LogF("WARNING: unable to start background auto host cleanup process: %s\n", err)
		return nil
	}

	if err := cmd.Process.Release(); err != nil {
		logboek.Context(ctx).Warn().LogF("WARNING: unable to detach background auto host cleanup process: %s\n", err)
		return nil
	}

	return nil
}

func RunHostCleanup(ctx context.Context, backend container_backend.ContainerBackend, options HostCleanupOptions) error {
	if err := logboek.Context(ctx).LogProcess("Running GC for tmp data").DoError(func() error {
		if err := tmp_manager.RunGC(ctx, options.DryRun, backend); err != nil {
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

	allowedDockerStorageVolumeUsagePercentage := getOptionValueOrDefault(options.AllowedDockerStorageVolumeUsagePercentage, DefaultAllowedDockerStorageVolumeUsagePercentage)
	allowedDockerStorageVolumeUsageMarginPercentage := getOptionValueOrDefault(options.AllowedDockerStorageVolumeUsageMarginPercentage, DefaultAllowedDockerStorageVolumeUsageMarginPercentage)

	cleaner, err := NewLocalBackendCleaner(backend)
	if errors.Is(err, ErrUnsupportedContainerBackend) {
		// if cleaner not implemented, skip cleaning
		return nil
	} else if err != nil {
		return err
	}

	return logboek.Context(ctx).Default().LogProcess("Running GC for local %s backend", cleaner.BackendName()).DoError(func() error {
		err := cleaner.RunGC(ctx, RunGCOptions{
			AllowedStorageVolumeUsagePercentage:       allowedDockerStorageVolumeUsagePercentage,
			AllowedStorageVolumeUsageMarginPercentage: allowedDockerStorageVolumeUsageMarginPercentage,
			StoragePath: *options.DockerServerStoragePath,
			Force:       options.Force,
			DryRun:      options.DryRun,
		})
		if err != nil {
			return fmt.Errorf("local %s backend GC failed: %w", cleaner.BackendName(), err)
		}
		return nil
	})
}

func ShouldRunAutoHostCleanup(ctx context.Context, backend container_backend.ContainerBackend, options HostCleanupOptions) (bool, error) {
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

	allowedDockerStorageVolumeUsagePercentage := getOptionValueOrDefault(options.AllowedDockerStorageVolumeUsagePercentage, DefaultAllowedDockerStorageVolumeUsagePercentage)

	cleaner, err := NewLocalBackendCleaner(backend)
	if errors.Is(err, ErrUnsupportedContainerBackend) {
		// if cleaner not implemented, skip cleaning
		return false, nil
	} else if err != nil {
		return false, err
	}

	shouldRun, err = cleaner.ShouldRunAutoGC(ctx, RunAutoGCOptions{
		AllowedStorageVolumeUsagePercentage: allowedDockerStorageVolumeUsagePercentage,
		StoragePath:                         *options.DockerServerStoragePath,
	})
	if err != nil {
		return false, fmt.Errorf("failed to check local docker server host cleaner GC: %w", err)
	}
	if shouldRun {
		return true, nil
	}

	return false, nil
}
