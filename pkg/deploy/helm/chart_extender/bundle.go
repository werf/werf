package chart_extender

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/postrender"
	"helm.sh/helm/v3/pkg/registry"
	"sigs.k8s.io/yaml"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/deploy/helm"
	"github.com/werf/werf/pkg/deploy/helm/chart_extender/helpers"
	"github.com/werf/werf/pkg/deploy/helm/command_helpers"
)

type BundleOptions struct {
	BuildChartDependenciesOpts        command_helpers.BuildChartDependenciesOptions
	ExtraAnnotations                  map[string]string
	ExtraLabels                       map[string]string
	IgnoreInvalidAnnotationsAndLabels bool
}

func NewBundle(ctx context.Context, dir string, helmEnvSettings *cli.EnvSettings, registryClient *registry.Client, opts BundleOptions) (*Bundle, error) {
	bundle := &Bundle{
		Dir:                            dir,
		HelmEnvSettings:                helmEnvSettings,
		RegistryClient:                 registryClient,
		BuildChartDependenciesOpts:     opts.BuildChartDependenciesOpts,
		ChartExtenderServiceValuesData: helpers.NewChartExtenderServiceValuesData(),
		ChartExtenderContextData:       helpers.NewChartExtenderContextData(ctx),
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

	bundle.extraAnnotationsAndLabelsPostRenderer = extraAnnotationsAndLabelsPostRenderer

	return bundle, nil
}

/*
 * Bundle object is chart.ChartExtender compatible object
 * which could be used during helm install/upgrade process
 */
type Bundle struct {
	Dir                        string
	HelmChart                  *chart.Chart
	HelmEnvSettings            *cli.EnvSettings
	RegistryClient             *registry.Client
	BuildChartDependenciesOpts command_helpers.BuildChartDependenciesOptions

	extraAnnotationsAndLabelsPostRenderer *helm.ExtraAnnotationsAndLabelsPostRenderer

	*helpers.ChartExtenderServiceValuesData
	*helpers.ChartExtenderContextData
}

func (bundle *Bundle) ChainPostRenderer(postRenderer postrender.PostRenderer) postrender.PostRenderer {
	var chain []postrender.PostRenderer

	if postRenderer != nil {
		chain = append(chain, postRenderer)
	}

	chain = append(chain, bundle.extraAnnotationsAndLabelsPostRenderer)

	return helm.NewPostRendererChain(chain...)
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

	res, err := LoadChartDependencies(bundle.ChartExtenderContext, func(ctx context.Context, dir string) ([]*chart.ChartExtenderBufferedFile, error) {
		files, err := loader.GetFilesFromLocalFilesystem(dir)
		if err != nil {
			return nil, err
		}
		return convertBufferedFilesForChartExtender(files), nil
	}, dir, convertBufferedFilesForChartExtender(files), bundle.HelmEnvSettings, bundle.RegistryClient, bundle.BuildChartDependenciesOpts)
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
