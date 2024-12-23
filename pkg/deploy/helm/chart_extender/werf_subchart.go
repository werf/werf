package chart_extender

import (
	"context"
	"text/template"

	"github.com/werf/3p-helm/pkg/chart"
	"github.com/werf/3p-helm/pkg/cli"
	"github.com/werf/3p-helm/pkg/werf/file"
	"github.com/werf/werf/v2/pkg/deploy/helm/chart_extender/helpers"
)

var _ chart.ChartExtender = (*WerfSubchart)(nil)

// NOTE: maybe in the future we will need a support for the werf project to be used as a chart.
// NOTE: This extender allows to define this behaviour.

type WerfSubchartOptions struct {
	DisableDefaultSecretValues bool
}

func NewWerfSubchart(
	ctx context.Context,
	opts WerfSubchartOptions,
) *WerfSubchart {
	return &WerfSubchart{
		ChartExtenderContextData:   helpers.NewChartExtenderContextData(ctx),
		DisableDefaultSecretValues: opts.DisableDefaultSecretValues,
	}
}

type WerfSubchart struct {
	HelmChart *chart.Chart

	DisableDefaultSecretValues bool

	*helpers.ChartExtenderContextData
}

// ChartCreated method for the chart.Extender interface
func (wc *WerfSubchart) ChartCreated(c *chart.Chart) error {
	wc.HelmChart = c
	return nil
}

// ChartLoaded method for the chart.Extender interface
func (wc *WerfSubchart) ChartLoaded(files []*file.ChartExtenderBufferedFile) error {
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
	return inputVals, nil
}

// SetupTemplateFuncs method for the chart.Extender interface
func (wc *WerfSubchart) SetupTemplateFuncs(t *template.Template, funcMap template.FuncMap) {
}

// LoadDir method for the chart.Extender interface
func (wc *WerfSubchart) LoadDir(dir string) (bool, []*file.ChartExtenderBufferedFile, error) {
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

func (wc *WerfSubchart) Type() string {
	return "subchart"
}

func (wc *WerfSubchart) GetChartFileReader() file.ChartFileReader {
	panic("not implemented")
}

func (wc *WerfSubchart) GetDisableDefaultSecretValues() bool {
	return wc.DisableDefaultSecretValues
}

func (wc *WerfSubchart) GetSecretValueFiles() []string {
	panic("not implemented")
}

func (wc *WerfSubchart) GetServiceValues() map[string]interface{} {
	return nil
}

func (wc *WerfSubchart) GetProjectDir() string {
	panic("not implemented")
}

func (wc *WerfSubchart) GetChartDir() string {
	panic("not implemented")
}
