package chart_extender

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/werf/3p-helm/pkg/chart"
	"github.com/werf/3p-helm/pkg/werf/file"
	"github.com/werf/werf/v2/pkg/deploy/helm/chart_extender/helpers"
)

var _ chart.ChartExtender = (*Bundle)(nil)

type BundleOptions struct {
	SecretValueFiles                  []string
	BuildChartDependenciesOpts        chart.BuildChartDependenciesOptions
	ExtraAnnotations                  map[string]string
	ExtraLabels                       map[string]string
	IgnoreInvalidAnnotationsAndLabels bool
	DisableDefaultValues              bool
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

	if dataMap, err := readBundleJsonMap(filepath.Join(bundle.Dir, "extra_annotations.json")); err != nil {
		return nil, err
	} else {
		bundle.AddExtraAnnotations(dataMap)
	}

	if dataMap, err := readBundleJsonMap(filepath.Join(bundle.Dir, "extra_labels.json")); err != nil {
		return nil, err
	} else {
		bundle.AddExtraLabels(dataMap)
	}

	bundle.AddExtraAnnotations(opts.ExtraAnnotations)
	bundle.AddExtraLabels(opts.ExtraLabels)

	return bundle, nil
}

/*
 * Bundle object is chart.ChartExtender compatible object
 * which could be used during helm install/upgrade process
 */
type Bundle struct {
	Dir                        string
	SecretValueFiles           []string
	HelmChart                  *chart.Chart
	BuildChartDependenciesOpts chart.BuildChartDependenciesOptions
	DisableDefaultValues       bool

	extraAnnotations map[string]string
	extraLabels      map[string]string

	*helpers.ChartExtenderServiceValuesData
}

func (bundle *Bundle) SetHelmChart(c *chart.Chart) {
	bundle.HelmChart = c
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

func (bundle *Bundle) GetDisableDefaultSecretValues() bool {
	panic("not implemented")
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

func (bundle *Bundle) AddExtraAnnotations(annotations map[string]string) {
	if bundle.extraAnnotations == nil {
		bundle.extraAnnotations = make(map[string]string)
	}

	for k, v := range annotations {
		bundle.extraAnnotations[k] = v
	}
}

func (bundle *Bundle) AddExtraLabels(labels map[string]string) {
	if bundle.extraLabels == nil {
		bundle.extraLabels = make(map[string]string)
	}

	for k, v := range labels {
		bundle.extraLabels[k] = v
	}
}

func (bundle *Bundle) GetExtraAnnotations() map[string]string {
	return bundle.extraAnnotations
}

func (bundle *Bundle) GetExtraLabels() map[string]string {
	return bundle.extraLabels
}

func writeBundleJsonMap(dataMap map[string]string, path string) error {
	if data, err := json.Marshal(dataMap); err != nil {
		return fmt.Errorf("unable to prepare %q data: %w", path, err)
	} else if err := ioutil.WriteFile(path, append(data, []byte("\n")...), os.ModePerm); err != nil {
		return fmt.Errorf("unable to write %q: %w", path, err)
	} else {
		return nil
	}
}

func readBundleJsonMap(path string) (map[string]string, error) {
	var res map[string]string
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("error accessing %q: %w", path, err)
	} else if data, err := ioutil.ReadFile(path); err != nil {
		return nil, fmt.Errorf("error reading %q: %w", path, err)
	} else if err := json.Unmarshal(data, &res); err != nil {
		return nil, fmt.Errorf("error unmarshalling json from %q: %w", path, err)
	} else {
		return res, nil
	}
}
