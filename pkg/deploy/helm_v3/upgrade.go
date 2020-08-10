package helm_v3

import (
	"context"
	"time"

	"helm.sh/helm/v3/pkg/cli/output"

	"helm.sh/helm/v3/pkg/cli/values"

	"github.com/werf/logboek"

	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/storage/driver"
)

type UpgradeOptions struct {
	ValuesOptions   values.Options
	Namespace       string
	CreateNamespace bool
	Install         bool
	Atomic          bool
	Timeout         time.Duration

	StatusProgressPeriod      time.Duration
	HooksStatusProgressPeriod time.Duration
}

func Upgrade(ctx context.Context, chart, releaseName string, opts UpgradeOptions) error {
	outfmt := output.Table

	envSettings := NewEnvSettings(opts.Namespace)
	cfg := NewActionConfig(envSettings, InitActionConfigOptions{StatusProgressPeriod: opts.StatusProgressPeriod, HooksStatusProgressPeriod: opts.HooksStatusProgressPeriod})
	client := action.NewUpgrade(cfg)

	client.Namespace = opts.Namespace
	client.Timeout = opts.Timeout
	client.Atomic = opts.Atomic
	client.Install = opts.Install

	// Fixes #7002 - Support reading values from STDIN for `upgrade` command
	// Must load values AFTER determining if we have to call install so that values loaded from stdin are are not read twice
	if client.Install {
		// If a release does not exist, install it.
		histClient := action.NewHistory(cfg)
		histClient.Max = 1
		if _, err := histClient.Run(releaseName); err == driver.ErrReleaseNotFound {
			logboek.Debug.LogF("Release %q does not exist. Installing it now.\n", releaseName)

			return Install(ctx, chart, releaseName, InstallOptions{
				ValuesOptions:   opts.ValuesOptions,
				Namespace:       opts.Namespace,
				CreateNamespace: opts.CreateNamespace,
				Install:         opts.Install,
				Atomic:          opts.Atomic,
				Timeout:         opts.Timeout,
			})
		} else if err != nil {
			return err
		}
	}

	if client.Version == "" && client.Devel {
		logboek.Debug.LogF("setting version to >0.0.0-0")
		client.Version = ">0.0.0-0"
	}

	chartPath, err := client.ChartPathOptions.LocateChart(chart, envSettings)
	if err != nil {
		return err
	}

	vals, err := opts.ValuesOptions.MergeValues(getter.All(envSettings))
	if err != nil {
		return err
	}

	// Check chart dependencies to make sure all are present in /charts
	ch, err := loader.Load(chartPath)
	if err != nil {
		return err
	}
	if req := ch.Metadata.Dependencies; req != nil {
		if err := action.CheckDependencies(ch, req); err != nil {
			return err
		}
	}

	rel, err := client.Run(releaseName, ch, vals)
	if err != nil {
		return errors.Wrap(err, "UPGRADE FAILED")
	}

	logboek.Debug.LogF("Release %q has been upgraded\n", releaseName)

	return outfmt.Write(logboek.GetOutStream(), &statusPrinter{rel, envSettings.Debug})
}
