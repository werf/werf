package chart_extender

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"helm.sh/helm/v3/pkg/cli"

	"github.com/werf/werf/pkg/deploy/helm"

	"helm.sh/helm/v3/pkg/chart"
)

func NewBundle(dir string) *Bundle {
	return &Bundle{
		Dir: dir,
	}
}

/*
 * Bundle object is chart.ChartExtender compatible object
 * which could be used during helm install/upgrade process
 */
type Bundle struct {
	Dir       string
	HelmChart *chart.Chart
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
	return inputVals, nil
}

// SetupTemplateFuncs method for the chart.Extender interface
func (bundle *Bundle) SetupTemplateFuncs(t *template.Template, funcMap template.FuncMap) {
	helmIncludeFunc := funcMap["include"].(func(name string, data interface{}) (string, error))
	setupIncludeWrapperFunc := func(name string) {
		funcMap[name] = func(data interface{}) (string, error) {
			return helmIncludeFunc(name, data)
		}
	}

	for _, name := range []string{"werf_image"} {
		setupIncludeWrapperFunc(name)
	}
}

// LoadDir method for the chart.Extender interface
func (bundle *Bundle) LoadDir(dir string) (bool, []*chart.ChartExtenderBufferedFile, error) {
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
