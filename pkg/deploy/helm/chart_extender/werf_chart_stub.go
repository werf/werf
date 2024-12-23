package chart_extender

import (
	"context"
	"text/template"

	"github.com/werf/3p-helm/pkg/chart"
	"github.com/werf/3p-helm/pkg/chartutil"
	"github.com/werf/3p-helm/pkg/cli"
	"github.com/werf/3p-helm/pkg/postrender"
	"github.com/werf/3p-helm/pkg/werf/file"
	"github.com/werf/werf/v2/pkg/deploy/helm"
	"github.com/werf/werf/v2/pkg/deploy/helm/chart_extender/helpers"
)

var _ chart.ChartExtender = (*WerfChartStub)(nil)

func NewWerfChartStub(ctx context.Context, ignoreInvalidAnnotationsAndLabels bool) *WerfChartStub {
	return &WerfChartStub{
		extraAnnotationsAndLabelsPostRenderer: helm.NewExtraAnnotationsAndLabelsPostRenderer(nil, nil, ignoreInvalidAnnotationsAndLabels),
		ChartExtenderContextData:              helpers.NewChartExtenderContextData(ctx),
	}
}

type WerfChartStub struct {
	HelmChart *chart.Chart
	ChartDir  string

	extraAnnotationsAndLabelsPostRenderer *helm.ExtraAnnotationsAndLabelsPostRenderer
	stubServiceValuesOverrides            map[string]interface{}
	stubServiceValues                     map[string]interface{}

	*helpers.ChartExtenderContextData
}

func (wc *WerfChartStub) AddExtraAnnotationsAndLabels(extraAnnotations, extraLabels map[string]string) {
	wc.extraAnnotationsAndLabelsPostRenderer.Add(extraAnnotations, extraLabels)
}

func (wc *WerfChartStub) ChainPostRenderer(postRenderer postrender.PostRenderer) postrender.PostRenderer {
	var chain []postrender.PostRenderer

	if postRenderer != nil {
		chain = append(chain, postRenderer)
	}

	chain = append(chain, wc.extraAnnotationsAndLabelsPostRenderer)

	return helm.NewPostRendererChain(chain...)
}

// ChartCreated method for the chart.Extender interface
func (wc *WerfChartStub) ChartCreated(c *chart.Chart) error {
	wc.HelmChart = c
	return nil
}

// ChartLoaded method for the chart.Extender interface
func (wc *WerfChartStub) ChartLoaded(files []*file.ChartExtenderBufferedFile) error {
	var opts helpers.GetHelmChartMetadataOptions
	opts.DefaultName = "stubchartname"
	opts.DefaultVersion = "1.0.0"
	wc.HelmChart.Metadata = helpers.AutosetChartMetadata(wc.HelmChart.Metadata, opts)

	wc.HelmChart.Templates = append(wc.HelmChart.Templates, &chart.File{
		Name: "templates/_werf_helpers.tpl",
	})

	return nil
}

// ChartDependenciesLoaded method for the chart.Extender interface
func (wc *WerfChartStub) ChartDependenciesLoaded() error {
	return nil
}

// MakeValues method for the chart.Extender interface
func (wc *WerfChartStub) MakeValues(_ map[string]interface{}) (
	map[string]interface{},
	error,
) {
	vals := make(map[string]interface{})
	chartutil.CoalesceTables(vals, wc.stubServiceValuesOverrides)
	chartutil.CoalesceTables(vals, wc.stubServiceValues)

	return vals, nil
}

// SetupTemplateFuncs method for the chart.Extender interface
func (wc *WerfChartStub) SetupTemplateFuncs(t *template.Template, funcMap template.FuncMap) {
}

// LoadDir method for the chart.Extender interface
func (wc *WerfChartStub) LoadDir(dir string) (bool, []*file.ChartExtenderBufferedFile, error) {
	wc.ChartDir = dir
	return false, nil, nil
}

// LocateChart method for the chart.Extender interface
func (wc *WerfChartStub) LocateChart(name string, settings *cli.EnvSettings) (bool, string, error) {
	return false, "", nil
}

// ReadFile method for the chart.Extender interface
func (wc *WerfChartStub) ReadFile(filePath string) (bool, []byte, error) {
	return false, nil, nil
}

func (wc *WerfChartStub) SetStubServiceValues(vals map[string]interface{}) {
	wc.stubServiceValues = vals
}

func (wc *WerfChartStub) SetStubServiceValuesOverrides(vals map[string]interface{}) {
	wc.stubServiceValuesOverrides = vals
}

func (wc *WerfChartStub) Type() string {
	return "chartstub"
}

func (wc *WerfChartStub) GetChartFileReader() file.ChartFileReader {
	panic("not implemented")
}

func (wc *WerfChartStub) GetDisableDefaultSecretValues() bool {
	panic("not implemented")
}

func (wc *WerfChartStub) GetSecretValueFiles() []string {
	return []string{}
}

func (wc *WerfChartStub) GetServiceValues() map[string]interface{} {
	panic("not implemented")
}

func (wc *WerfChartStub) GetProjectDir() string {
	panic("not implemented")
}

func (wc *WerfChartStub) GetChartDir() string {
	return wc.ChartDir
}
