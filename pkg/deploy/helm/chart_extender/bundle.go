package chart_extender

import (
	"context"

	"github.com/werf/3p-helm/pkg/chart"
	"github.com/werf/3p-helm/pkg/werf/file"
	"github.com/werf/werf/v2/pkg/deploy/helm/chart_extender/helpers"
)

var _ chart.ChartExtender = (*Bundle)(nil)

type BundleOptions struct {
	SecretValueFiles           []string
	BuildChartDependenciesOpts chart.BuildChartDependenciesOptions
	DisableDefaultValues       bool
}

func NewBundle(
	ctx context.Context,
	dir string,
	opts BundleOptions,
) (*Bundle, error) {
	bundle := &Bundle{
		Dir:                            dir,
		SecretValueFiles:               opts.SecretValueFiles,
		BuildChartDependenciesOpts:     opts.BuildChartDependenciesOpts,
		ChartExtenderServiceValuesData: helpers.NewChartExtenderServiceValuesData(),
		DisableDefaultValues:           opts.DisableDefaultValues,
	}

	return bundle, nil
}

/*
 * Bundle object is chart.ChartExtender compatible object
 * which could be used during helm install/upgrade process
 */
type Bundle struct {
	Dir                        string
	SecretValueFiles           []string
	BuildChartDependenciesOpts chart.BuildChartDependenciesOptions
	DisableDefaultValues       bool

	*helpers.ChartExtenderServiceValuesData
}

func (bundle *Bundle) Type() string {
	return "bundle"
}

func (bundle *Bundle) GetChartFileReader() file.ChartFileReader {
	panic("not implemented")
}

func (bundle *Bundle) GetDisableDefaultValues() bool {
	return bundle.DisableDefaultValues
}

func (bundle *Bundle) GetSecretValueFiles() []string {
	return bundle.SecretValueFiles
}

func (bundle *Bundle) GetProjectDir() string {
	panic("not implemented")
}

func (bundle *Bundle) GetChartDir() string {
	return bundle.Dir
}

func (bundle *Bundle) SetChartDir(dir string) {
	bundle.Dir = dir
}

func (bundle *Bundle) GetBuildChartDependenciesOpts() chart.BuildChartDependenciesOptions {
	return bundle.BuildChartDependenciesOpts
}
