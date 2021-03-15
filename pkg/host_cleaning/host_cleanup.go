package host_cleaning

import (
	"context"
	"fmt"

	"github.com/werf/lockgate"
	"github.com/werf/werf/pkg/werf"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/tmp_manager"
)

type HostCleanupOptions struct {
	DryRun                                bool
	AllowedVolumeUsagePercentageThreshold int64
	Force                                 bool
	DockerServerStoragePath               string
}

func HostCleanup(ctx context.Context, options HostCleanupOptions) error {
	return werf.WithHostLock(ctx, "gc", lockgate.AcquireOptions{}, func() error {
		if err := tmp_manager.GC(ctx, options.DryRun); err != nil {
			return fmt.Errorf("tmp files GC failed: %s", err)
		}

		return logboek.Context(ctx).Default().LogProcess("Running GC for local docker server").DoError((func() error {
			return RunGCForLocalDockerServer(ctx, LocalDockerServerGCOptions{
				AllowedVolumeUsagePercentageThreshold: options.AllowedVolumeUsagePercentageThreshold,
				DryRun:                                options.DryRun,
				Force:                                 options.Force,
				DockerServerStoragePath:               options.DockerServerStoragePath,
			})
		}))
	})
}
