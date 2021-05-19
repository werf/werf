package chart_extender

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/mitchellh/copystructure"
	"github.com/werf/werf/pkg/deploy/secrets_manager"

	"github.com/werf/logboek"
	"sigs.k8s.io/yaml"

	helm_v3 "helm.sh/helm/v3/cmd/helm"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"

	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/deploy/helm"
	"github.com/werf/werf/pkg/deploy/helm/command_helpers"
	"github.com/werf/werf/pkg/giterminism_manager"

	"github.com/werf/werf/pkg/deploy/helm/chart_extender/helpers"
	"github.com/werf/werf/pkg/deploy/helm/chart_extender/helpers/secrets"
)

type WerfChartOptions struct {
	SecretValueFiles           []string
	ExtraAnnotations           map[string]string
	ExtraLabels                map[string]string
	BuildChartDependenciesOpts command_helpers.BuildChartDependenciesOptions
	DisableSecrets             bool
}

func NewWerfChart(ctx context.Context, giterminismManager giterminism_manager.Interface, secretsManager *secrets_manager.SecretsManager, chartDir string, helmEnvSettings *cli.EnvSettings, registryClientHandle *helm_v3.RegistryClientHandle, opts WerfChartOptions) *WerfChart {
	wc := &WerfChart{
		ChartDir:             chartDir,
		SecretValueFiles:     opts.SecretValueFiles,
		HelmEnvSettings:      helmEnvSettings,
		RegistryClientHandle: registryClientHandle,
		DisableSecrets:       opts.DisableSecrets,

		GiterminismManager: giterminismManager,
		SecretsManager:     secretsManager,

		extraAnnotationsAndLabelsPostRenderer: helm.NewExtraAnnotationsAndLabelsPostRenderer(nil, nil),

		ChartExtenderServiceValuesData: helpers.NewChartExtenderServiceValuesData(),
		ChartExtenderContextData:       helpers.NewChartExtenderContextData(ctx),
	}

	wc.extraAnnotationsAndLabelsPostRenderer.Add(opts.ExtraAnnotations, opts.ExtraLabels)

	return wc
}

type WerfChartRuntimeData struct {
	DecodedSecretValues    map[string]interface{}
	DecodedSecretFilesData map[string]string
	SecretValuesToMask     []string
}

type WerfChart struct {
	HelmChart *chart.Chart

	ChartDir                   string
	SecretValueFiles           []string
	HelmEnvSettings            *cli.EnvSettings
	RegistryClientHandle       *helm_v3.RegistryClientHandle
	BuildChartDependenciesOpts command_helpers.BuildChartDependenciesOptions
	DisableSecrets             bool

	GiterminismManager giterminism_manager.Interface
	SecretsManager     *secrets_manager.SecretsManager

	extraAnnotationsAndLabelsPostRenderer *helm.ExtraAnnotationsAndLabelsPostRenderer
	werfConfig                            *config.WerfConfig

	*secrets.SecretsRuntimeData
	*helpers.ChartExtenderServiceValuesData
	*helpers.ChartExtenderContextData
}

// ChartCreated method for the chart.Extender interface
func (wc *WerfChart) ChartCreated(c *chart.Chart) error {
	wc.HelmChart = c
	wc.SecretsRuntimeData = secrets.NewSecretsRuntimeData()
	return nil
}

// ChartLoaded method for the chart.Extender interface
func (wc *WerfChart) ChartLoaded(files []*chart.ChartExtenderBufferedFile) error {
	if wc.SecretsManager != nil {
		if err := wc.SecretsRuntimeData.DecodeAndLoadSecrets(wc.ChartExtenderContext, files, wc.ChartDir, wc.GiterminismManager.ProjectDir(), wc.SecretsManager, secrets.DecodeAndLoadSecretsOptions{
			GiterminismManager:     wc.GiterminismManager,
			CustomSecretValueFiles: wc.SecretValueFiles,
		}); err != nil {
			return fmt.Errorf("error decoding secrets: %s", err)
		}
	}

	var opts helpers.GetHelmChartMetadataOptions
	if wc.werfConfig != nil {
		opts.OverrideName = wc.werfConfig.Meta.Project
	}
	opts.DefaultVersion = "1.0.0"
	wc.HelmChart.Metadata = helpers.AutosetChartMetadata(wc.HelmChart.Metadata, opts)

	wc.HelmChart.Templates = append(wc.HelmChart.Templates, &chart.File{
		Name: "templates/_werf_helpers.tpl",
		Data: []byte(helpers.ChartTemplateHelpers),
	})

	return nil
}

// ChartDependenciesLoaded method for the chart.Extender interface
func (wc *WerfChart) ChartDependenciesLoaded() error {
	return nil
}

func (wc *WerfChart) makeValues(inputVals map[string]interface{}, withSecrets bool) (map[string]interface{}, error) {
	vals := make(map[string]interface{})

	chartutil.CoalesceTables(vals, wc.ServiceValues) // NOTE: service values will not be saved into the marshalled release

	if withSecrets {
		chartutil.CoalesceTables(vals, wc.SecretsRuntimeData.DecodedSecretValues)
	}

	chartutil.CoalesceTables(vals, inputVals)

	data, err := yaml.Marshal(vals)
	logboek.Context(wc.ChartExtenderContext).Debug().LogF("-- WerfChart.makeValues result (err=%v):\n%s\n---\n", err, data)

	return vals, nil
}

// MakeValues method for the chart.Extender interface
func (wc *WerfChart) MakeValues(inputVals map[string]interface{}) (map[string]interface{}, error) {
	return wc.makeValues(inputVals, true)
}

func (wc *WerfChart) MakeBundleValues(chrt *chart.Chart, inputVals map[string]interface{}) (map[string]interface{}, error) {
	vals, err := wc.makeValues(inputVals, false)
	if err != nil {
		return nil, fmt.Errorf("failed to coalesce werf chart values: %s", err)
	}

	v, err := copystructure.Copy(vals)
	if err != nil {
		return vals, err
	}

	valsCopy := v.(map[string]interface{})
	// if we have an empty map, make sure it is initialized
	if valsCopy == nil {
		valsCopy = make(map[string]interface{})
	}

	chartutil.CoalesceChartValues(chrt, valsCopy)

	data, err := yaml.Marshal(vals)
	logboek.Context(wc.ChartExtenderContext).Debug().LogF("-- WerfChart.MakeBundleValues result (err=%v):\n%s\n---\n", err, data)

	return valsCopy, nil
}

// SetupTemplateFuncs method for the chart.Extender interface
func (wc *WerfChart) SetupTemplateFuncs(t *template.Template, funcMap template.FuncMap) {
	funcMap["werf_secret_file"] = func(secretRelativePath string) (string, error) {
		if path.IsAbs(secretRelativePath) {
			return "", fmt.Errorf("expected relative secret file path, given path %v", secretRelativePath)
		}

		decodedData, ok := wc.SecretsRuntimeData.DecodedSecretFilesData[secretRelativePath]

		if !ok {
			var secretFiles []string
			for key := range wc.SecretsRuntimeData.DecodedSecretFilesData {
				secretFiles = append(secretFiles, key)
			}

			return "", fmt.Errorf("secret file %q not found, you may use one of the following: %q", secretRelativePath, strings.Join(secretFiles, "', '"))
		}

		return decodedData, nil
	}

	helpers.SetupIncludeWrapperFuncs(funcMap)
	helpers.SetupWerfImageDeprecationFunc(wc.ChartExtenderContext, funcMap)
}

// LoadDir method for the chart.Extender interface
func (wc *WerfChart) LoadDir(dir string) (bool, []*chart.ChartExtenderBufferedFile, error) {
	files, err := wc.GiterminismManager.FileReader().LoadChartDir(wc.ChartExtenderContext, dir)
	if err != nil {
		return true, nil, fmt.Errorf("giterministic files loader failed: %s", err)
	}

	res, err := LoadChartDependencies(wc.ChartExtenderContext, files, wc.HelmEnvSettings, wc.RegistryClientHandle, wc.BuildChartDependenciesOpts)
	if err != nil {
		return true, res, fmt.Errorf("chart dependencies loader failed: %s", err)
	}
	return true, res, err
}

// LocateChart method for the chart.Extender interface
func (wc *WerfChart) LocateChart(name string, settings *cli.EnvSettings) (bool, string, error) {
	res, err := wc.GiterminismManager.FileReader().LocateChart(wc.ChartExtenderContext, name, settings)
	return true, res, err
}

// ReadFile method for the chart.Extender interface
func (wc *WerfChart) ReadFile(filePath string) (bool, []byte, error) {
	res, err := wc.GiterminismManager.FileReader().ReadChartFile(wc.ChartExtenderContext, filePath)
	return true, res, err
}

func (wc *WerfChart) GetPostRenderer() (*helm.ExtraAnnotationsAndLabelsPostRenderer, error) {
	return wc.extraAnnotationsAndLabelsPostRenderer, nil
}

func (wc *WerfChart) SetWerfConfig(werfConfig *config.WerfConfig) error {
	wc.extraAnnotationsAndLabelsPostRenderer.Add(map[string]string{
		"project.werf.io/name": werfConfig.Meta.Project,
	}, nil)

	wc.werfConfig = werfConfig

	return nil
}

func (wc *WerfChart) SetEnv(env string) error {
	wc.extraAnnotationsAndLabelsPostRenderer.Add(map[string]string{
		"project.werf.io/env": env,
	}, nil)

	return nil
}

/*
 * CreateNewBundle creates new Bundle object with werf chart extensions taken into account.
 * inputVals could contain any custom values, which should be stored in the bundle.
 */
func (wc *WerfChart) CreateNewBundle(ctx context.Context, destDir string, inputVals map[string]interface{}) (*Bundle, error) {
	if err := os.RemoveAll(destDir); err != nil {
		return nil, fmt.Errorf("unable to remove %q: %s", destDir, err)
	}
	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to create dir %q: %s", destDir, err)
	}

	chartPath := filepath.Join(wc.GiterminismManager.ProjectDir(), wc.ChartDir)
	chrt, err := loader.LoadDir(chartPath)
	if err != nil {
		return nil, fmt.Errorf("error loading chart %q: %s", chartPath, err)
	}

	vals, err := wc.MakeBundleValues(chrt, inputVals)
	if err != nil {
		return nil, fmt.Errorf("unable to construct bundle input values: %s", err)
	}

	valsData, err := json.MarshalIndent(vals, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("unable to prepare values: %s", err)
	}

	logboek.Context(ctx).Debug().LogF("Saving bundle values:\n%s\n---\n", valsData)

	valuesFile := filepath.Join(destDir, "values.yaml")
	if err := ioutil.WriteFile(valuesFile, append(valsData, []byte("\n")...), os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to write %q: %s", valuesFile, err)
	}

	if wc.HelmChart.Metadata == nil {
		panic("unexpected condition")
	}

	bundleMetadata := *wc.HelmChart.Metadata
	// Force api v2
	bundleMetadata.APIVersion = chart.APIVersionV2

	chartYamlFile := filepath.Join(destDir, "Chart.yaml")
	if data, err := json.Marshal(bundleMetadata); err != nil {
		return nil, fmt.Errorf("unable to prepare Chart.yaml data: %s", err)
	} else if err := ioutil.WriteFile(chartYamlFile, append(data, []byte("\n")...), os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to write %q: %s", chartYamlFile, err)
	}

	if wc.HelmChart.Lock != nil {
		chartLockFile := filepath.Join(destDir, "Chart.lock")
		if data, err := json.Marshal(wc.HelmChart.Lock); err != nil {
			return nil, fmt.Errorf("unable to prepare Chart.lock data: %s", err)
		} else if err := ioutil.WriteFile(chartLockFile, append(data, []byte("\n")...), os.ModePerm); err != nil {
			return nil, fmt.Errorf("unable to write %q: %s", chartLockFile, err)
		}
	}

	templatesDir := filepath.Join(destDir, "templates")
	if err := os.MkdirAll(templatesDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to create dir %q: %s", templatesDir, err)
	}

	for _, f := range wc.HelmChart.Templates {
		p := filepath.Join(destDir, f.Name)
		dir := filepath.Dir(p)

		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return nil, fmt.Errorf("error creating dir %q: %s", dir, err)
		}

		if err := ioutil.WriteFile(p, append(f.Data, []byte("\n")...), os.ModePerm); err != nil {
			return nil, fmt.Errorf("unable to write %q: %s", p, err)
		}
	}

	if wc.HelmChart.Schema != nil {
		schemaFile := filepath.Join(destDir, "values.schema.json")
		if data, err := json.Marshal(wc.HelmChart.Schema); err != nil {
			return nil, fmt.Errorf("unable to prepare values.schema.json data: %s", err)
		} else if err := ioutil.WriteFile(schemaFile, append(data, []byte("\n")...), os.ModePerm); err != nil {
			return nil, fmt.Errorf("unable to write %q: %s", schemaFile, err)
		}
	}

	if wc.extraAnnotationsAndLabelsPostRenderer.ExtraAnnotations != nil {
		if err := writeBundleJsonMap(wc.extraAnnotationsAndLabelsPostRenderer.ExtraAnnotations, filepath.Join(destDir, "extra_annotations.json")); err != nil {
			return nil, err
		}
	}

	if wc.extraAnnotationsAndLabelsPostRenderer.ExtraLabels != nil {
		if err := writeBundleJsonMap(wc.extraAnnotationsAndLabelsPostRenderer.ExtraLabels, filepath.Join(destDir, "extra_labels.json")); err != nil {
			return nil, err
		}
	}

	return NewBundle(ctx, destDir, wc.HelmEnvSettings, wc.RegistryClientHandle, BundleOptions{BuildChartDependenciesOpts: wc.BuildChartDependenciesOpts}), nil
}
