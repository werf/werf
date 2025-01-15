package chart_extender

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/werf/3p-helm/pkg/chart"
	"github.com/werf/3p-helm/pkg/cli"
	"github.com/werf/3p-helm/pkg/postrender"
	"github.com/werf/3p-helm/pkg/registry"
	"github.com/werf/3p-helm/pkg/werf/file"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/deploy/helm"
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
	helmEnvSettings *cli.EnvSettings,
	registryClient *registry.Client,
	opts BundleOptions,
) (*Bundle, error) {
	bundle := &Bundle{
		Dir:                            dir,
		SecretValueFiles:               opts.SecretValueFiles,
		HelmEnvSettings:                helmEnvSettings,
		RegistryClient:                 registryClient,
		BuildChartDependenciesOpts:     opts.BuildChartDependenciesOpts,
		ChartExtenderServiceValuesData: helpers.NewChartExtenderServiceValuesData(),
		DisableDefaultValues:           opts.DisableDefaultValues,
	}

	extraAnnotationsAndLabelsPostRenderer := helm.NewExtraAnnotationsAndLabelsPostRenderer(nil, nil, opts.IgnoreInvalidAnnotationsAndLabels)

	if dataMap, err := readBundleJsonMap(filepath.Join(bundle.Dir, "extra_annotations.json")); err != nil {
		return nil, err
	} else {
		extraAnnotationsAndLabelsPostRenderer.Add(dataMap, nil)
	}

	if dataMap, err := readBundleJsonMap(filepath.Join(bundle.Dir, "extra_labels.json")); err != nil {
		return nil, err
	} else {
		extraAnnotationsAndLabelsPostRenderer.Add(nil, dataMap)
	}

	extraAnnotationsAndLabelsPostRenderer.Add(opts.ExtraAnnotations, opts.ExtraLabels)

	bundle.ExtraAnnotationsAndLabelsPostRenderer = extraAnnotationsAndLabelsPostRenderer

	return bundle, nil
}

/*
 * Bundle object is chart.ChartExtender compatible object
 * which could be used during helm install/upgrade process
 */
type Bundle struct {
	Dir                                   string
	SecretValueFiles                      []string
	HelmChart                             *chart.Chart
	HelmEnvSettings                       *cli.EnvSettings
	RegistryClient                        *registry.Client
	BuildChartDependenciesOpts            chart.BuildChartDependenciesOptions
	DisableDefaultValues                  bool
	ExtraAnnotationsAndLabelsPostRenderer *helm.ExtraAnnotationsAndLabelsPostRenderer

	*helpers.ChartExtenderServiceValuesData
}

func (bundle *Bundle) ChainPostRenderer(postRenderer postrender.PostRenderer) postrender.PostRenderer {
	var chain []postrender.PostRenderer

	if postRenderer != nil {
		chain = append(chain, postRenderer)
	}

	chain = append(chain, bundle.ExtraAnnotationsAndLabelsPostRenderer)

	return helm.NewPostRendererChain(chain...)
}

// ChartCreated method for the chart.Extender interface
func (bundle *Bundle) ChartCreated(c *chart.Chart) error {
	bundle.HelmChart = c
	return nil
}

// ChartLoaded method for the chart.Extender interface
func (bundle *Bundle) ChartLoaded(files []*file.ChartExtenderBufferedFile) error {
	if bundle.DisableDefaultValues {
		logboek.Context(context.Background()).Info().LogF("Disable default werf chart values\n")
		bundle.HelmChart.Values = nil
	}

	return nil
}

// ChartDependenciesLoaded method for the chart.Extender interface
func (bundle *Bundle) ChartDependenciesLoaded() error {
	return nil
}

// MakeValues method for the chart.Extender interface
func (bundle *Bundle) MakeValues(inputVals map[string]interface{}) (map[string]interface{}, error) {
	return inputVals, nil
}

// SetupTemplateFuncs method for the chart.Extender interface
func (bundle *Bundle) SetupTemplateFuncs(t *template.Template, funcMap template.FuncMap) {
}

// LoadDir method for the chart.Extender interface
func (bundle *Bundle) LoadDir(dir string) (bool, []*file.ChartExtenderBufferedFile, error) {
	return false, nil, nil
}

// LocateChart method for the chart.Extender interface
func (bundle *Bundle) LocateChart(name string, settings *cli.EnvSettings) (bool, string, error) {
	return false, "", nil
}

// ReadFile method for the chart.Extender interface
func (bundle *Bundle) ReadFile(filePath string) (bool, []byte, error) {
	return false, nil, nil
}

func (bundle *Bundle) Type() string {
	return "bundle"
}

func (bundle *Bundle) GetChartFileReader() file.ChartFileReader {
	panic("not implemented")
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
