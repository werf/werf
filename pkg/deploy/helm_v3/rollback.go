package helm_v3

import (
	"fmt"
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
}

func Rollback(releaseName string, opts RollbackOptions) error {
	return logboek.Default.LogProcess(fmt.Sprintf("Rolling back release %q"), logboek.LevelLogProcessOptions{}, func() error {
		envSettings := NewEnvSettings(opts.Namespace)
		cfg := NewActionConfig(envSettings)
		client := action.NewRollback(cfg)
		client.Timeout = opts.Timeout
		client.Version = opts.Version

		if err := client.Run(releaseName); err != nil {
			return err
		}

		return nil
	})
}
