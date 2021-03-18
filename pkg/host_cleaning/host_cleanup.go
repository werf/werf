package host_cleaning

import (
	"context"
	"fmt"
	"time"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/tmp_manager"
	"github.com/werf/werf/pkg/werf"
)

type HostCleanupOptions struct {
	AllowedVolumeUsagePercentage       *uint
	AllowedVolumeUsageMarginPercentage *uint

	DryRun                  bool
	Force                   bool
	DockerServerStoragePath string
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
		logboek.Context(ctx).Default().LogFDetails(" - least recently used werf images (only >= v1.2 werf images could be removed, note that werf <= v1.1 images will not be deleted by this auto cleanup);\n")
		logboek.Context(ctx).Default().LogLn()
		logboek.Context(ctx).Default().LogFDetails("To disable this auto host cleanup please specify --disable-auto-host-cleanup option (or WERF_DISABLE_AUTO_HOST_CLEANUP=true environment variable).\n")
		logboek.Context(ctx).Default().LogFDetails("Host cleanup could be performed manually with the `werf host cleanup` command, please set this command into crontab for your host in the case when auto host cleanup disabled.\n")
		logboek.Context(ctx).Default().LogLn()

		return RunHostCleanup(ctx, options)
	})
}

func RunHostCleanup(ctx context.Context, options HostCleanupOptions) error {
	if err := tmp_manager.GC(ctx, options.DryRun); err != nil {
		return fmt.Errorf("tmp files GC failed: %s", err)
	}

	return logboek.Context(ctx).Default().LogProcess("Running GC for local docker server").DoError((func() error {
		return RunGCForLocalDockerServer(ctx, options)
	}))
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

	shouldRun, err = ShouldRunAutoGCForLocalDockerServer(ctx, options)
	if err != nil {
		return false, fmt.Errorf("failed to check local docker server host cleaner GC: %s", err)
	}
	return shouldRun, nil
}
