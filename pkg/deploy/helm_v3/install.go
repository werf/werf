package helm_v3

import (
	"context"
	"time"

	"helm.sh/helm/v3/pkg/cli/output"

	"github.com/werf/logboek"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli/values"

	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
)

type InstallOptions struct {
	ValuesOptions   values.Options
	Namespace       string
	CreateNamespace bool
	Install         bool
	Atomic          bool
	Timeout         time.Duration

	StatusProgressPeriod      time.Duration
	HooksStatusProgressPeriod time.Duration
}

func Install(ctx context.Context, chart, releaseName string, opts InstallOptions) error {
	outfmt := output.Table

	envSettings := NewEnvSettings(opts.Namespace)
	cfg := NewActionConfig(envSettings, InitActionConfigOptions{StatusProgressPeriod: opts.StatusProgressPeriod, HooksStatusProgressPeriod: opts.HooksStatusProgressPeriod})
	client := action.NewInstall(cfg)
	client.Namespace = opts.Namespace
	client.CreateNamespace = opts.CreateNamespace
	client.ReleaseName = releaseName

	logboek.Debug.LogF("Original chart version: %q", client.Version)
	if client.Version == "" && client.Devel {
		logboek.Debug.LogF("setting version to >0.0.0-0")
		client.Version = ">0.0.0-0"
	}

	cp, err := client.ChartPathOptions.LocateChart(chart, envSettings)
	if err != nil {
		return err
	}

	logboek.Debug.LogF("CHART PATH: %s\n", cp)

	p := getter.All(envSettings)
	vals, err := opts.ValuesOptions.MergeValues(p)
	if err != nil {
		return err
	}

	// Check chart dependencies to make sure all are present in /charts
	chartRequested, err := loader.Load(cp)
	if err != nil {
		return err
	}

	validInstallableChart, err := isChartInstallable(chartRequested)
	if !validInstallableChart {
		return err
	}

	if req := chartRequested.Metadata.Dependencies; req != nil {
		// If CheckDependencies returns an error, we have unfulfilled dependencies.
		// As of Helm 2.4.0, this is treated as a stopping condition:
		// https://github.com/helm/helm/issues/2209
		if err := action.CheckDependencies(chartRequested, req); err != nil {
			if client.DependencyUpdate {
				man := &downloader.Manager{
					Out:              logboek.GetOutStream(),
					ChartPath:        cp,
					Keyring:          client.ChartPathOptions.Keyring,
					SkipUpdate:       false,
					Getters:          p,
					RepositoryConfig: envSettings.RepositoryConfig,
					RepositoryCache:  envSettings.RepositoryCache,
					Debug:            envSettings.Debug,
				}
				if err := man.Update(); err != nil {
					return err
				}
				// Reload the chart with the updated Chart.lock file.
				if chartRequested, err = loader.Load(cp); err != nil {
					return errors.Wrap(err, "failed reloading chart after repo update")
				}
			} else {
				return err
			}
		}
	}

	rel, err := client.Run(chartRequested, vals)
	if err != nil {
		return errors.Wrap(err, "INSTALL FAILED")
	}

	return outfmt.Write(logboek.GetOutStream(), &statusPrinter{rel, envSettings.Debug})
}

// isChartInstallable validates if a chart can be installed
//
// Application chart type is only installable
func isChartInstallable(ch *chart.Chart) (bool, error) {
	switch ch.Metadata.Type {
	case "", "application":
		return true, nil
	}
	return false, errors.Errorf("%s charts are not installable", ch.Metadata.Type)
}
