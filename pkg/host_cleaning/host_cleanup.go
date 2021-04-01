package host_cleaning

import (
	"context"
	"fmt"
	"time"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/git_repo/gitdata"
	"github.com/werf/werf/pkg/tmp_manager"
	"github.com/werf/werf/pkg/werf"
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

	DryRun                  bool
	Force                   bool
	DockerServerStoragePath string
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

func RunAutoHostCleanup(ctx context.Context, options HostCleanupOptions) error {
	shouldRun, err := ShouldRunAutoHostCleanup(ctx, options)
	if err != nil {
		return err
	}
	if !shouldRun {
		return nil
	}

	return logboek.Context(ctx).Default().LogProcess("Running auto host cleanup").DoError(func() error {
		logboek.Context(ctx).Default().LogFDetails("### Auto host cleanup note ###\n")
		logboek.Context(ctx).Default().LogLn()
		logboek.Context(ctx).Default().LogFDetails("Werf tries to maintain host clean by deleting:\n")
		logboek.Context(ctx).Default().LogFDetails(" - old unused files from werf caches (which are stored in the ~/.werf/local_cache);\n")
		logboek.Context(ctx).Default().LogFDetails(" - old temporary service files /tmp/werf-project-data-* and /tmp/werf-config-render-*;\n")
		logboek.Context(ctx).Default().LogFDetails(" - least recently used werf images;\n")
		logboek.Context(ctx).Default().LogLn()
		logboek.Context(ctx).Default().LogFDetails("NOTE: Werf-host-cleanup procedure of v1.2 werf version will not cleanup --stages-storage=:local stages of v1.1 werf version, because this is primary stages storage data, and it can only be cleaned by the regular per-project werf-cleanup command with git-history based algorithm.\n")
		logboek.Context(ctx).Default().LogLn()
		logboek.Context(ctx).Default().LogFDetails("To disable this auto host cleanup please specify --disable-auto-host-cleanup option (or WERF_DISABLE_AUTO_HOST_CLEANUP=true environment variable).\n")
		logboek.Context(ctx).Default().LogFDetails("Host cleanup could be performed manually with the `werf host cleanup` command, please set this command into crontab for your host in the case when auto host cleanup disabled.\n")
		logboek.Context(ctx).Default().LogLn()

		return RunHostCleanup(ctx, options)
	})
}

func RunHostCleanup(ctx context.Context, options HostCleanupOptions) error {
	if err := logboek.Context(ctx).LogProcess("Running GC for tmp data").DoError(func() error {
		if err := tmp_manager.RunGC(ctx, options.DryRun); err != nil {
			return fmt.Errorf("tmp files GC failed: %s", err)
		}
		return nil
	}); err != nil {
		return err
	}

	allowedLocalCacheVolumeUsagePercentage := getOptionValueOrDefault(options.AllowedLocalCacheVolumeUsagePercentage, DefaultAllowedLocalCacheVolumeUsagePercentage)
	allowedLocalCacheVolumeUsageMarginPercentage := getOptionValueOrDefault(options.AllowedLocalCacheVolumeUsageMarginPercentage, DefaultAllowedLocalCacheVolumeUsageMarginPercentage)

	allowedDockerStorageVolumeUsagePercentage := getOptionValueOrDefault(options.AllowedDockerStorageVolumeUsagePercentage, DefaultAllowedDockerStorageVolumeUsagePercentage)
	allowedDockerStorageVolumeUsageMarginPercentage := getOptionValueOrDefault(options.AllowedDockerStorageVolumeUsageMarginPercentage, DefaultAllowedDockerStorageVolumeUsageMarginPercentage)

	if err := logboek.Context(ctx).Default().LogProcess("Running GC for git data").DoError(func() error {
		if err := gitdata.RunGC(ctx, allowedLocalCacheVolumeUsagePercentage, allowedLocalCacheVolumeUsageMarginPercentage); err != nil {
			return fmt.Errorf("git repo GC failed: %s", err)
		}
		return nil
	}); err != nil {
		return err
	}

	dockerServerStoragePath, err := getDockerServerStoragePath(ctx, options.DockerServerStoragePath)
	if err != nil {
		return fmt.Errorf("error getting local docker server storage path: %s", err)
	}

	return logboek.Context(ctx).Default().LogProcess("Running GC for local docker server").DoError(func() error {
		if err := RunGCForLocalDockerServer(ctx, allowedDockerStorageVolumeUsagePercentage, allowedDockerStorageVolumeUsageMarginPercentage, dockerServerStoragePath, options.Force, options.DryRun); err != nil {
			return fmt.Errorf("local docker server GC failed: %s", err)
		}
		return nil
	})
}

func ShouldRunAutoHostCleanup(ctx context.Context, options HostCleanupOptions) (bool, error) {
	t, err := werf.GetWerfFirstRunAt(ctx)
	if err != nil {
		return false, fmt.Errorf("error getting last werf run timestamp: %s", err)
	}
	// Only run auto host cleanup on persistent hosts
	if t.IsZero() || time.Since(t) <= 2*time.Hour {
		return false, nil
	}

	shouldRun, err := tmp_manager.ShouldRunAutoGC()
	if err != nil {
		return false, fmt.Errorf("failed to check tmp manager GC: %s", err)
	}
	if shouldRun {
		return true, nil
	}

	allowedLocalCacheVolumeUsagePercentage := getOptionValueOrDefault(options.AllowedLocalCacheVolumeUsagePercentage, DefaultAllowedLocalCacheVolumeUsagePercentage)
	allowedDockerStorageVolumeUsagePercentage := getOptionValueOrDefault(options.AllowedDockerStorageVolumeUsagePercentage, DefaultAllowedDockerStorageVolumeUsagePercentage)

	shouldRun, err = gitdata.ShouldRunAutoGC(ctx, allowedLocalCacheVolumeUsagePercentage)
	if err != nil {
		return false, fmt.Errorf("failed to check git repo GC: %s", err)
	}
	if shouldRun {
		return true, nil
	}

	dockerServerStoragePath, err := getDockerServerStoragePath(ctx, options.DockerServerStoragePath)
	if err != nil {
		return false, fmt.Errorf("error getting local docker server storage path: %s", err)
	}

	shouldRun, err = ShouldRunAutoGCForLocalDockerServer(ctx, allowedDockerStorageVolumeUsagePercentage, dockerServerStoragePath)
	if err != nil {
		return false, fmt.Errorf("failed to check local docker server host cleaner GC: %s", err)
	}
	return shouldRun, nil
}
