package chart_extender

import (
	"text/template"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
)

// NOTE: maybe in the future we will need a support for the werf project to be used as a chart.
// NOTE: This extender allows to define this behaviour.

func NewWerfSubchart() *WerfSubchart {
	return &WerfSubchart{}
}

type WerfSubchart struct {
	HelmChart *chart.Chart
}

// ChartCreated method for the chart.Extender interface
func (wc *WerfSubchart) ChartCreated(c *chart.Chart) error {
	wc.HelmChart = c
	return nil
}

// ChartLoaded method for the chart.Extender interface
func (wc *WerfSubchart) ChartLoaded(files []*chart.ChartExtenderBufferedFile) error {
	return nil
}

// ChartDependenciesLoaded method for the chart.Extender interface
func (wc *WerfSubchart) ChartDependenciesLoaded() error {
	return nil
}

// MakeValues method for the chart.Extender interface
func (wc *WerfSubchart) MakeValues(inputVals map[string]interface{}) (map[string]interface{}, error) {
	vals := make(map[string]interface{})
	chartutil.CoalesceTables(vals, inputVals)
	return vals, nil
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
