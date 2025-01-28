package chart_extender

import (
	"context"

	"github.com/werf/3p-helm/pkg/chart"
	"github.com/werf/3p-helm/pkg/werf/file"
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
		DisableDefaultSecretValues: opts.DisableDefaultSecretValues,
	}
}

type WerfSubchart struct {
	HelmChart *chart.Chart

	DisableDefaultSecretValues bool
}

// SetHelmChart method for the chart.Extender interface
func (wc *WerfSubchart) SetHelmChart(c *chart.Chart) {
	wc.HelmChart = c
}

func (wc *WerfSubchart) Type() string {
	return "subchart"
}

func (wc *WerfSubchart) GetChartFileReader() file.ChartFileReader {
	panic("not implemented")
}

func (wc *WerfSubchart) GetDisableDefaultValues() bool {
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

func (wc *WerfSubchart) SetChartDir(dir string) {
	panic("not implemented")
}

func (wc *WerfSubchart) GetBuildChartDependenciesOpts() chart.BuildChartDependenciesOptions {
	panic("not implemented")
}

func (wc *WerfSubchart) AddExtraAnnotations(annotations map[string]string) {
	panic("not implemented")
}

func (wc *WerfSubchart) AddExtraLabels(labels map[string]string) {
	panic("not implemented")
}

func (wc *WerfSubchart) GetExtraAnnotations() map[string]string {
	panic("not implemented")
}

func (wc *WerfSubchart) GetExtraLabels() map[string]string {
	panic("not implemented")
}
