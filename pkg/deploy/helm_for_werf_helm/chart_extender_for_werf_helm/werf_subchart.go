package chart_extender_for_werf_helm

import (
	"context"
	"fmt"
	"text/template"

	chart "github.com/werf/3p-helm-for-werf-helm/pkg/chart"
	cli "github.com/werf/3p-helm-for-werf-helm/pkg/cli"
	"github.com/werf/logboek"
	secrets_manager "github.com/werf/nelm-for-werf-helm/pkg/secrets_manager"
	helpers "github.com/werf/werf/v2/pkg/deploy/helm_for_werf_helm/chart_extender_for_werf_helm/helpers_for_werf_helm"
	secrets "github.com/werf/werf/v2/pkg/deploy/helm_for_werf_helm/chart_extender_for_werf_helm/helpers_for_werf_helm/secrets_for_werf_helm"
)

// NOTE: maybe in the future we will need a support for the werf project to be used as a chart.
// NOTE: This extender allows to define this behavior.

type WerfSubchartOptions struct {
	DisableDefaultSecretValues bool
}

func NewWerfSubchart(
	ctx context.Context,
	secretsManager *secrets_manager.SecretsManager,
	opts WerfSubchartOptions,
) *WerfSubchart {
	return &WerfSubchart{
		SecretsManager:             secretsManager,
		ChartExtenderContextData:   helpers.NewChartExtenderContextData(ctx),
		DisableDefaultSecretValues: opts.DisableDefaultSecretValues,
	}
}

type WerfSubchart struct {
	HelmChart      *chart.Chart
	SecretsManager *secrets_manager.SecretsManager

	DisableDefaultSecretValues bool

	*secrets.SecretsRuntimeData
	*helpers.ChartExtenderContextData
	*helpers.ChartExtenderValuesMerger
}

// ChartCreated method for the chart.Extender interface
func (wc *WerfSubchart) ChartCreated(c *chart.Chart) error {
	wc.HelmChart = c
	wc.SecretsRuntimeData = secrets.NewSecretsRuntimeData()
	return nil
}

// ChartLoaded method for the chart.Extender interface
func (wc *WerfSubchart) ChartLoaded(files []*chart.ChartExtenderBufferedFile) error {
	if wc.SecretsManager != nil {
		if wc.DisableDefaultSecretValues {
			logboek.Context(wc.ChartExtenderContext).Info().LogF("Disabled subchart secret values\n")
		}

		if err := wc.SecretsRuntimeData.DecodeAndLoadSecrets(wc.ChartExtenderContext, files, "", "", wc.SecretsManager, secrets.DecodeAndLoadSecretsOptions{
			WithoutDefaultSecretValues: wc.DisableDefaultSecretValues,
		}); err != nil {
			return fmt.Errorf("error decoding secrets: %w", err)
		}
	}

	return nil
}

// ChartDependenciesLoaded method for the chart.Extender interface
func (wc *WerfSubchart) ChartDependenciesLoaded() error {
	return nil
}

// MakeValues method for the chart.Extender interface
func (wc *WerfSubchart) MakeValues(inputVals map[string]interface{}) (
	map[string]interface{},
	error,
) {
	return wc.MergeValues(wc.ChartExtenderContext, inputVals, nil, wc.SecretsRuntimeData)
}

// SetupTemplateFuncs method for the chart.Extender interface
func (wc *WerfSubchart) SetupTemplateFuncs(t *template.Template, funcMap template.FuncMap) {
}

// LoadDir method for the chart.Extender interface
func (wc *WerfSubchart) LoadDir(dir string) (bool, []*chart.ChartExtenderBufferedFile, error) {
	return false, nil, nil
}

// LocateChart method for the chart.Extender interface
func (wc *WerfSubchart) LocateChart(name string, settings *cli.EnvSettings) (bool, string, error) {
	return false, "", nil
}

// ReadFile method for the chart.Extender interface
func (wc *WerfSubchart) ReadFile(filePath string) (bool, []byte, error) {
	return false, nil, nil
}
