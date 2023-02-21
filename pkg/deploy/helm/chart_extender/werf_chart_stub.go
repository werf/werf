package chart_extender

import (
	"context"
	"fmt"
	"os"
	"text/template"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/postrender"

	"github.com/werf/werf/pkg/deploy/helm"
	"github.com/werf/werf/pkg/deploy/helm/chart_extender/helpers"
	"github.com/werf/werf/pkg/deploy/helm/chart_extender/helpers/secrets"
	"github.com/werf/werf/pkg/deploy/secrets_manager"
)

func NewWerfChartStub(ctx context.Context, ignoreInvalidAnnotationsAndLabels bool) *WerfChartStub {
	return &WerfChartStub{
		extraAnnotationsAndLabelsPostRenderer: helm.NewExtraAnnotationsAndLabelsPostRenderer(nil, nil, ignoreInvalidAnnotationsAndLabels),
		ChartExtenderContextData:              helpers.NewChartExtenderContextData(ctx),
	}
}

type WerfChartStub struct {
	HelmChart        *chart.Chart
	ChartDir         string
	SecretsManager   *secrets_manager.SecretsManager
	SecretValueFiles []string

	extraAnnotationsAndLabelsPostRenderer *helm.ExtraAnnotationsAndLabelsPostRenderer
	stubServiceValuesOverrides            map[string]interface{}
	stubServiceValues                     map[string]interface{}

	*secrets.SecretsRuntimeData
	*helpers.ChartExtenderContextData
}

func (wc *WerfChartStub) SetupSecretsManager(manager *secrets_manager.SecretsManager) {
	wc.SecretsManager = manager
}

func (wc *WerfChartStub) AddExtraAnnotationsAndLabels(extraAnnotations, extraLabels map[string]string) {
	wc.extraAnnotationsAndLabelsPostRenderer.Add(extraAnnotations, extraLabels)
}

func (wc *WerfChartStub) SetupSecretValueFiles(secretValueFiles []string) {
	wc.SecretValueFiles = secretValueFiles
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
	wc.SecretsRuntimeData = secrets.NewSecretsRuntimeData()
	return nil
}

// ChartLoaded method for the chart.Extender interface
func (wc *WerfChartStub) ChartLoaded(files []*chart.ChartExtenderBufferedFile) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current process working directory: %w", err)
	}

	if wc.SecretsManager != nil {
		if err := wc.SecretsRuntimeData.DecodeAndLoadSecrets(wc.ChartExtenderContext, files, wc.ChartDir, cwd, wc.SecretsManager, secrets.DecodeAndLoadSecretsOptions{
			CustomSecretValueFiles:  wc.SecretValueFiles,
			LoadFromLocalFilesystem: true,
		}); err != nil {
			return fmt.Errorf("error decoding secrets: %w", err)
		}
	}

	var opts helpers.GetHelmChartMetadataOptions
	opts.DefaultName = "stubchartname"
	opts.DefaultVersion = "1.0.0"
	wc.HelmChart.Metadata = helpers.AutosetChartMetadata(wc.HelmChart.Metadata, opts)

	wc.HelmChart.Templates = append(wc.HelmChart.Templates, &chart.File{
		Name: "templates/_werf_helpers.tpl",
		Data: []byte(helpers.ChartTemplateHelpers),
	})

	return nil
}

// ChartDependenciesLoaded method for the chart.Extender interface
func (wc *WerfChartStub) ChartDependenciesLoaded() error {
	return nil
}

// MakeValues method for the chart.Extender interface
func (wc *WerfChartStub) MakeValues(inputVals map[string]interface{}) (map[string]interface{}, error) {
	vals := make(map[string]interface{})
	chartutil.CoalesceTables(vals, wc.stubServiceValuesOverrides)
	chartutil.CoalesceTables(vals, wc.stubServiceValues)
	chartutil.CoalesceTables(vals, wc.SecretsRuntimeData.DecryptedSecretValues)
	chartutil.CoalesceTables(vals, inputVals)

	return vals, nil
}

// SetupTemplateFuncs method for the chart.Extender interface
func (wc *WerfChartStub) SetupTemplateFuncs(t *template.Template, funcMap template.FuncMap) {
	helpers.SetupWerfSecretFile(wc.SecretsRuntimeData, funcMap)
	helpers.SetupIncludeWrapperFuncs(funcMap)
	helpers.SetupWerfImageDeprecationFunc(wc.ChartExtenderContext, funcMap)
}

// LoadDir method for the chart.Extender interface
func (wc *WerfChartStub) LoadDir(dir string) (bool, []*chart.ChartExtenderBufferedFile, error) {
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
