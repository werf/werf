package helm_v3

import (
	"context"
	"time"

	"github.com/werf/logboek"
	"helm.sh/helm/v3/pkg/action"
)

type RollbackOptions struct {
	Namespace string
	Version   int

	DryRun        bool
	Recreate      bool
	Force         bool
	DisableHooks  bool
	Timeout       time.Duration
	Wait          bool
	CleanupOnFail bool

	StatusProgressPeriod      time.Duration
	HooksStatusProgressPeriod time.Duration
}

func Rollback(ctx context.Context, releaseName string, opts RollbackOptions) error {
	return logboek.Context(ctx).Default().LogProcess("Rolling back release %q", releaseName).DoError(func() error {
		envSettings := NewEnvSettings(ctx, opts.Namespace)
		cfg := NewActionConfig(ctx, envSettings, InitActionConfigOptions{StatusProgressPeriod: opts.StatusProgressPeriod, HooksStatusProgressPeriod: opts.HooksStatusProgressPeriod})
		client := action.NewRollback(cfg)
		client.Timeout = opts.Timeout
		client.Version = opts.Version

		return client.Run(releaseName)
	})
}
