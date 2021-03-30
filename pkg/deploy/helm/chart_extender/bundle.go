package chart_extender

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/werf/logboek"
	"sigs.k8s.io/yaml"

	helm_v3 "helm.sh/helm/v3/cmd/helm"
	"helm.sh/helm/v3/pkg/chartutil"

	"github.com/werf/werf/pkg/deploy/helm/chart_extender/helpers"
	"github.com/werf/werf/pkg/deploy/helm/command_helpers"

	"helm.sh/helm/v3/pkg/chart/loader"

	"helm.sh/helm/v3/pkg/cli"

	"github.com/werf/werf/pkg/deploy/helm"

	"helm.sh/helm/v3/pkg/chart"
)

type BundleOptions struct {
	BuildChartDependenciesOpts command_helpers.BuildChartDependenciesOptions
}

func NewBundle(ctx context.Context, dir string, helmEnvSettings *cli.EnvSettings, registryClientHandle *helm_v3.RegistryClientHandle, opts BundleOptions) *Bundle {
	return &Bundle{
		Dir:                            dir,
		HelmEnvSettings:                helmEnvSettings,
		RegistryClientHandle:           registryClientHandle,
		BuildChartDependenciesOpts:     opts.BuildChartDependenciesOpts,
		ChartExtenderServiceValuesData: helpers.NewChartExtenderServiceValuesData(),
		ChartExtenderContextData:       helpers.NewChartExtenderContextData(ctx),
	}
}

/*
 * Bundle object is chart.ChartExtender compatible object
 * which could be used during helm install/upgrade process
 */
type Bundle struct {
	Dir                        string
	HelmChart                  *chart.Chart
	HelmEnvSettings            *cli.EnvSettings
	RegistryClientHandle       *helm_v3.RegistryClientHandle
	BuildChartDependenciesOpts command_helpers.BuildChartDependenciesOptions

	*helpers.ChartExtenderServiceValuesData
	*helpers.ChartExtenderContextData
}

func (bundle *Bundle) GetPostRenderer() (*helm.ExtraAnnotationsAndLabelsPostRenderer, error) {
	postRenderer := helm.NewExtraAnnotationsAndLabelsPostRenderer(nil, nil)

	if dataMap, err := readBundleJsonMap(filepath.Join(bundle.Dir, "extra_annotations.json")); err != nil {
		return nil, err
	} else {
		postRenderer.Add(dataMap, nil)
	}

	if dataMap, err := readBundleJsonMap(filepath.Join(bundle.Dir, "extra_labels.json")); err != nil {
		return nil, err
	} else {
		postRenderer.Add(nil, dataMap)
	}

	return postRenderer, nil
}

// ChartCreated method for the chart.Extender interface
func (bundle *Bundle) ChartCreated(c *chart.Chart) error {
	bundle.HelmChart = c
	return nil
}

// ChartLoaded method for the chart.Extender interface
func (bundle *Bundle) ChartLoaded(files []*chart.ChartExtenderBufferedFile) error {
	return nil
}

// ChartDependenciesLoaded method for the chart.Extender interface
func (bundle *Bundle) ChartDependenciesLoaded() error {
	return nil
}

// MakeValues method for the chart.Extender interface
func (bundle *Bundle) MakeValues(inputVals map[string]interface{}) (map[string]interface{}, error) {
	vals := make(map[string]interface{})

	chartutil.CoalesceTables(vals, bundle.ServiceValues)
	chartutil.CoalesceTables(vals, inputVals)

	data, err := yaml.Marshal(vals)
	logboek.Context(bundle.ChartExtenderContext).Debug().LogF("-- Bundle.MakeValues result (err=%v):\n%s\n---\n", err, data)

	return vals, nil
}

// SetupTemplateFuncs method for the chart.Extender interface
func (bundle *Bundle) SetupTemplateFuncs(t *template.Template, funcMap template.FuncMap) {
	helpers.SetupIncludeWrapperFuncs(funcMap)
	helpers.SetupWerfImageDeprecationFunc(bundle.ChartExtenderContext, funcMap)
}

func convertBufferedFilesForChartExtender(files []*loader.BufferedFile) []*chart.ChartExtenderBufferedFile {
	var res []*chart.ChartExtenderBufferedFile
	for _, f := range files {
		f1 := new(chart.ChartExtenderBufferedFile)
		*f1 = chart.ChartExtenderBufferedFile(*f)
		res = append(res, f1)
	}
	return res
}

// LoadDir method for the chart.Extender interface
func (bundle *Bundle) LoadDir(dir string) (bool, []*chart.ChartExtenderBufferedFile, error) {
	files, err := loader.GetFilesFromLocalFilesystem(dir)
	if err != nil {
		return true, nil, err
	}

	res, err := LoadChartDependencies(bundle.ChartExtenderContext, convertBufferedFilesForChartExtender(files), bundle.HelmEnvSettings, bundle.RegistryClientHandle, bundle.BuildChartDependenciesOpts)
	return true, res, err
}

// LocateChart method for the chart.Extender interface
func (bundle *Bundle) LocateChart(name string, settings *cli.EnvSettings) (bool, string, error) {
	return false, "", nil
}

// ReadFile method for the chart.Extender interface
func (bundle *Bundle) ReadFile(filePath string) (bool, []byte, error) {
	return false, nil, nil
}

func writeBundleJsonMap(dataMap map[string]string, path string) error {
	if data, err := json.Marshal(dataMap); err != nil {
		return fmt.Errorf("unable to prepare %q data: %s", path, err)
	} else if err := ioutil.WriteFile(path, append(data, []byte("\n")...), os.ModePerm); err != nil {
		return fmt.Errorf("unable to write %q: %s", path, err)
	} else {
		return nil
	}
}

func readBundleJsonMap(path string) (map[string]string, error) {
	var res map[string]string
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("error accessing %q: %s", path, err)
	} else if data, err := ioutil.ReadFile(path); err != nil {
		return nil, fmt.Errorf("error reading %q: %s", path, err)
	} else if err := json.Unmarshal(data, &res); err != nil {
		return nil, fmt.Errorf("error unmarshalling json from %q: %s", path, err)
	} else {
		return res, nil
	}
}
