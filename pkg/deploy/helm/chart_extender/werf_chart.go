package chart_extender

import (
	"context"

	"github.com/werf/3p-helm/pkg/chart"
	"github.com/werf/3p-helm/pkg/werf/file"
	"github.com/werf/werf/v2/pkg/deploy/helm/chart_extender/helpers"
)

var _ chart.ChartExtender = (*WerfChart)(nil)

type WerfChartOptions struct {
	SecretValueFiles                  []string
	ExtraAnnotations                  map[string]string
	ExtraLabels                       map[string]string
	BuildChartDependenciesOpts        chart.BuildChartDependenciesOptions
	IgnoreInvalidAnnotationsAndLabels bool
	DisableDefaultValues              bool
	DisableDefaultSecretValues        bool
}

func NewWerfChart(
	ctx context.Context,
	chartFileReader file.ChartFileReader,
	chartDir string,
	projectDir string,
	opts WerfChartOptions,
) *WerfChart {
	wc := &WerfChart{
		ChartDir:         chartDir,
		ProjectDir:       projectDir,
		SecretValueFiles: opts.SecretValueFiles,

		ChartFileReader: chartFileReader,

		ChartExtenderServiceValuesData: helpers.NewChartExtenderServiceValuesData(),

		DisableDefaultValues:       opts.DisableDefaultValues,
		DisableDefaultSecretValues: opts.DisableDefaultSecretValues,
		BuildChartDependenciesOpts: opts.BuildChartDependenciesOpts,
	}

	wc.AddExtraAnnotations(opts.ExtraAnnotations)
	wc.AddExtraLabels(opts.ExtraLabels)

	return wc
}

type WerfChart struct {
	HelmChart *chart.Chart

	ChartDir                   string
	ProjectDir                 string
	SecretValueFiles           []string
	BuildChartDependenciesOpts chart.BuildChartDependenciesOptions
	DisableDefaultValues       bool
	DisableDefaultSecretValues bool

	ChartFileReader file.ChartFileReader

	extraAnnotations map[string]string
	extraLabels      map[string]string

	*helpers.ChartExtenderServiceValuesData
}

// SetHelmChart method for the chart.Extender interface
func (wc *WerfChart) SetHelmChart(c *chart.Chart) {
	wc.HelmChart = c
}

func (wc *WerfChart) AddExtraAnnotations(annotations map[string]string) {
	if wc.extraAnnotations == nil {
		wc.extraAnnotations = make(map[string]string)
	}

	for k, v := range annotations {
		wc.extraAnnotations[k] = v
	}
}

func (wc *WerfChart) AddExtraLabels(labels map[string]string) {
	if wc.extraLabels == nil {
		wc.extraLabels = make(map[string]string)
	}

	for k, v := range labels {
		wc.extraLabels[k] = v
	}
}

func (wc *WerfChart) GetExtraAnnotations() map[string]string {
	return wc.extraAnnotations
}

func (wc *WerfChart) GetExtraLabels() map[string]string {
	return wc.extraLabels
}

func (wc *WerfChart) Type() string {
	return "chart"
}

func (wc *WerfChart) GetChartFileReader() file.ChartFileReader {
	return wc.ChartFileReader
}

func (wc *WerfChart) GetDisableDefaultValues() bool {
	return wc.DisableDefaultValues
}

func (wc *WerfChart) GetDisableDefaultSecretValues() bool {
	return wc.DisableDefaultSecretValues
}

func (wc *WerfChart) GetSecretValueFiles() []string {
	return wc.SecretValueFiles
}

func (wc *WerfChart) GetProjectDir() string {
	return wc.ProjectDir
}

func (wc *WerfChart) GetChartDir() string {
	return wc.ChartDir
}

func (wc *WerfChart) SetChartDir(dir string) {
	wc.ChartDir = dir
}

func (wc *WerfChart) GetBuildChartDependenciesOpts() chart.BuildChartDependenciesOptions {
	return wc.BuildChartDependenciesOpts
}
