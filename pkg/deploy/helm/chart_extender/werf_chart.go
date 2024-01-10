package chart_extender

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/mitchellh/copystructure"
	helm_v3 "helm.sh/helm/v3/cmd/helm"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/postrender"
	"helm.sh/helm/v3/pkg/registry"
	"sigs.k8s.io/yaml"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/deploy/helm"
	"github.com/werf/werf/pkg/deploy/helm/chart_extender/helpers"
	"github.com/werf/werf/pkg/deploy/helm/chart_extender/helpers/secrets"
	"github.com/werf/werf/pkg/deploy/helm/command_helpers"
	"github.com/werf/werf/pkg/deploy/secrets_manager"
	"github.com/werf/werf/pkg/giterminism_manager"
	"github.com/werf/werf/pkg/util"
)

type WerfChartOptions struct {
	SecretValueFiles                  []string
	ExtraAnnotations                  map[string]string
	ExtraLabels                       map[string]string
	BuildChartDependenciesOpts        command_helpers.BuildChartDependenciesOptions
	IgnoreInvalidAnnotationsAndLabels bool
	DisableDefaultValues              bool
	DisableDefaultSecretValues        bool
}

func NewWerfChart(ctx context.Context, giterminismManager giterminism_manager.Interface, secretsManager *secrets_manager.SecretsManager, chartDir string, helmEnvSettings *cli.EnvSettings, registryClient *registry.Client, opts WerfChartOptions) *WerfChart {
	wc := &WerfChart{
		ChartDir:         chartDir,
		SecretValueFiles: opts.SecretValueFiles,
		HelmEnvSettings:  helmEnvSettings,
		RegistryClient:   registryClient,

		GiterminismManager: giterminismManager,
		SecretsManager:     secretsManager,

		extraAnnotationsAndLabelsPostRenderer: helm.NewExtraAnnotationsAndLabelsPostRenderer(nil, nil, opts.IgnoreInvalidAnnotationsAndLabels),

		ChartExtenderServiceValuesData: helpers.NewChartExtenderServiceValuesData(),
		ChartExtenderContextData:       helpers.NewChartExtenderContextData(ctx),

		DisableDefaultValues:       opts.DisableDefaultValues,
		DisableDefaultSecretValues: opts.DisableDefaultSecretValues,
		BuildChartDependenciesOpts: opts.BuildChartDependenciesOpts,
	}

	wc.extraAnnotationsAndLabelsPostRenderer.Add(opts.ExtraAnnotations, opts.ExtraLabels)

	return wc
}

type WerfChartRuntimeData struct {
	DecryptedSecretValues    map[string]interface{}
	DecryptedSecretFilesData map[string]string
	SecretValuesToMask       []string
}

type WerfChart struct {
	HelmChart *chart.Chart

	ChartDir                   string
	SecretValueFiles           []string
	HelmEnvSettings            *cli.EnvSettings
	RegistryClient             *registry.Client
	BuildChartDependenciesOpts command_helpers.BuildChartDependenciesOptions
	DisableDefaultValues       bool
	DisableDefaultSecretValues bool

	GiterminismManager giterminism_manager.Interface
	SecretsManager     *secrets_manager.SecretsManager

	extraAnnotationsAndLabelsPostRenderer *helm.ExtraAnnotationsAndLabelsPostRenderer
	werfConfig                            *config.WerfConfig

	*secrets.SecretsRuntimeData
	*helpers.ChartExtenderServiceValuesData
	*helpers.ChartExtenderContextData
	*helpers.ChartExtenderValuesMerger
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
		if wc.DisableDefaultSecretValues {
			logboek.Context(wc.ChartExtenderContext).Info().LogF("Disable default werf chart secret values\n")
		}

		if err := wc.SecretsRuntimeData.DecodeAndLoadSecrets(wc.ChartExtenderContext, files, wc.ChartDir, wc.GiterminismManager.ProjectDir(), wc.SecretsManager, secrets.DecodeAndLoadSecretsOptions{
			GiterminismManager:         wc.GiterminismManager,
			CustomSecretValueFiles:     wc.SecretValueFiles,
			WithoutDefaultSecretValues: wc.DisableDefaultSecretValues,
		}); err != nil {
			return fmt.Errorf("error decoding secrets: %w", err)
		}
	}

	var opts helpers.GetHelmChartMetadataOptions
	if wc.werfConfig != nil {
		opts.DefaultName = wc.werfConfig.Meta.Project
	}
	opts.DefaultVersion = "1.0.0"
	wc.HelmChart.Metadata = helpers.AutosetChartMetadata(wc.HelmChart.Metadata, opts)

	wc.HelmChart.Templates = append(wc.HelmChart.Templates, &chart.File{
		Name: "templates/_werf_helpers.tpl",
		Data: []byte(helpers.ChartTemplateHelpers),
	})

	if wc.DisableDefaultValues {
		logboek.Context(wc.ChartExtenderContext).Info().LogF("Disable default werf chart values\n")
		wc.HelmChart.Values = nil
	}

	return nil
}

// ChartDependenciesLoaded method for the chart.Extender interface
func (wc *WerfChart) ChartDependenciesLoaded() error {
	return nil
}

// MakeValues method for the chart.Extender interface
func (wc *WerfChart) MakeValues(inputVals map[string]interface{}) (map[string]interface{}, error) {
	return wc.MergeValues(wc.ChartExtenderContext, inputVals, wc.ServiceValues, wc.SecretsRuntimeData)
}

func (wc *WerfChart) MakeBundleValues(chrt *chart.Chart, inputVals map[string]interface{}) (map[string]interface{}, error) {
	helpers.DebugPrintValues(wc.ChartExtenderContext, "input", inputVals)

	vals, err := wc.MergeValues(wc.ChartExtenderContext, inputVals, wc.ServiceValues, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to coalesce werf chart values: %w", err)
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

	chartutil.CoalesceChartValues(chrt, valsCopy, true)

	helpers.DebugPrintValues(wc.ChartExtenderContext, "all", valsCopy)

	return valsCopy, nil
}

func (wc *WerfChart) MakeBundleSecretValues(ctx context.Context, secretsRuntimeData *secrets.SecretsRuntimeData) (map[string]interface{}, error) {
	if helpers.DebugSecretValues() {
		helpers.DebugPrintValues(wc.ChartExtenderContext, "secret", wc.SecretsRuntimeData.DecryptedSecretValues)
	}
	return secretsRuntimeData.GetEncodedSecretValues(ctx, wc.SecretsManager, wc.GiterminismManager.ProjectDir())
}

// SetupTemplateFuncs method for the chart.Extender interface
func (wc *WerfChart) SetupTemplateFuncs(t *template.Template, funcMap template.FuncMap) {
	helpers.SetupWerfSecretFile(wc.SecretsRuntimeData, funcMap)
	helpers.SetupIncludeWrapperFuncs(funcMap)
	helpers.SetupWerfImageDeprecationFunc(wc.ChartExtenderContext, funcMap)
}

// LoadDir method for the chart.Extender interface
func (wc *WerfChart) LoadDir(dir string) (bool, []*chart.ChartExtenderBufferedFile, error) {
	chartFiles, err := wc.GiterminismManager.FileReader().LoadChartDir(wc.ChartExtenderContext, dir)
	if err != nil {
		return true, nil, fmt.Errorf("giterministic files loader failed: %w", err)
	}

	res, err := LoadChartDependencies(wc.ChartExtenderContext, wc.GiterminismManager.FileReader().LoadChartDir, dir, chartFiles, wc.HelmEnvSettings, wc.RegistryClient, wc.BuildChartDependenciesOpts)
	if err != nil {
		return true, res, fmt.Errorf("chart dependencies loader failed: %w", err)
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

func (wc *WerfChart) ChainPostRenderer(postRenderer postrender.PostRenderer) postrender.PostRenderer {
	var chain []postrender.PostRenderer

	if postRenderer != nil {
		chain = append(chain, postRenderer)
	}

	chain = append(chain, wc.extraAnnotationsAndLabelsPostRenderer)

	return helm.NewPostRendererChain(chain...)
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
func (wc *WerfChart) CreateNewBundle(ctx context.Context, destDir, chartVersion string, vals *values.Options) (*Bundle, error) {
	chartPath := filepath.Join(wc.GiterminismManager.ProjectDir(), wc.ChartDir)
	chrt, err := loader.LoadDir(chartPath)
	if err != nil {
		return nil, fmt.Errorf("error loading chart %q: %w", chartPath, err)
	}

	var valsData []byte
	{
		p := getter.All(helm_v3.Settings)
		vals, err := vals.MergeValues(p, wc)
		if err != nil {
			return nil, fmt.Errorf("unable to merge input values: %w", err)
		}

		bundleVals, err := wc.MakeBundleValues(chrt, vals)
		if err != nil {
			return nil, fmt.Errorf("unable to construct bundle values: %w", err)
		}

		valsData, err = yaml.Marshal(bundleVals)
		if err != nil {
			return nil, fmt.Errorf("unable to marshal bundle values: %w", err)
		}
	}

	var secretValsData []byte
	if wc.SecretsRuntimeData != nil && !wc.SecretsManager.IsMissedSecretKeyModeEnabled() {
		vals, err := wc.MakeBundleSecretValues(ctx, wc.SecretsRuntimeData)
		if err != nil {
			return nil, fmt.Errorf("unable to construct bundle secret values: %w", err)
		}

		secretValsData, err = yaml.Marshal(vals)
		if err != nil {
			return nil, fmt.Errorf("unable to marshal bundle secret values: %w", err)
		}
	}

	if destDir == "" {
		destDir = wc.HelmChart.Metadata.Name
	}

	if err := os.RemoveAll(destDir); err != nil {
		return nil, fmt.Errorf("unable to remove %q: %w", destDir, err)
	}
	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to create dir %q: %w", destDir, err)
	}

	logboek.Context(ctx).Debug().LogF("Saving bundle values:\n%s\n---\n", valsData)

	valuesFile := filepath.Join(destDir, "values.yaml")
	if err := ioutil.WriteFile(valuesFile, valsData, os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to write %q: %w", valuesFile, err)
	}

	if secretValsData != nil {
		secretValuesFile := filepath.Join(destDir, "secret-values.yaml")
		if err := ioutil.WriteFile(secretValuesFile, secretValsData, os.ModePerm); err != nil {
			return nil, fmt.Errorf("unable to write %q: %w", secretValuesFile, err)
		}
	}

	if wc.HelmChart.Metadata == nil {
		panic("unexpected condition")
	}

	bundleMetadata := *wc.HelmChart.Metadata
	// Force api v2
	bundleMetadata.APIVersion = chart.APIVersionV2
	bundleMetadata.Version = chartVersion

	chartYamlFile := filepath.Join(destDir, "Chart.yaml")
	if data, err := json.Marshal(bundleMetadata); err != nil {
		return nil, fmt.Errorf("unable to prepare Chart.yaml data: %w", err)
	} else if err := ioutil.WriteFile(chartYamlFile, append(data, []byte("\n")...), os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to write %q: %w", chartYamlFile, err)
	}

	if wc.HelmChart.Lock != nil {
		chartLockFile := filepath.Join(destDir, "Chart.lock")
		if data, err := json.Marshal(wc.HelmChart.Lock); err != nil {
			return nil, fmt.Errorf("unable to prepare Chart.lock data: %w", err)
		} else if err := ioutil.WriteFile(chartLockFile, append(data, []byte("\n")...), os.ModePerm); err != nil {
			return nil, fmt.Errorf("unable to write %q: %w", chartLockFile, err)
		}
	}

	templatesDir := filepath.Join(destDir, "templates")
	if err := os.MkdirAll(templatesDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to create dir %q: %w", templatesDir, err)
	}

	for _, f := range wc.HelmChart.Templates {
		if err := writeChartFile(ctx, destDir, f.Name, f.Data); err != nil {
			return nil, fmt.Errorf("error writing chart template: %w", err)
		}
	}

	chartDirAbs := filepath.Join(wc.GiterminismManager.ProjectDir(), wc.ChartDir)

	ignoreChartValuesFiles := []string{secrets.DefaultSecretValuesFileName}

	// Do not publish into the bundle no custom values nor custom secret values.
	// Final bundle values and secret values will be preconstructed, merged and
	//  embedded into the bundle using only 2 files: values.yaml and secret-values.yaml.
	for _, customValuesPath := range append(wc.SecretValueFiles, vals.ValueFiles...) {
		path := util.GetAbsoluteFilepath(customValuesPath)
		if util.IsSubpathOfBasePath(chartDirAbs, path) {
			ignoreChartValuesFiles = append(ignoreChartValuesFiles, util.GetRelativeToBaseFilepath(chartDirAbs, path))
		}
	}

WritingFiles:
	for _, f := range wc.HelmChart.Files {
		for _, ignoreValuesFile := range ignoreChartValuesFiles {
			if f.Name == ignoreValuesFile {
				continue WritingFiles
			}
		}

		if err := writeChartFile(ctx, destDir, f.Name, f.Data); err != nil {
			return nil, fmt.Errorf("error writing miscellaneous chart file: %w", err)
		}
	}

	for _, dep := range wc.HelmChart.Metadata.Dependencies {
		var depPath string

		switch {
		case dep.Repository == "":
			depPath = filepath.Join("charts", dep.Name)
		case strings.HasPrefix(dep.Repository, "file://"):
			depPath = strings.TrimPrefix(dep.Repository, "file://")
		default:
			continue
		}

		for _, f := range wc.HelmChart.Raw {
			if strings.HasPrefix(f.Name, depPath) {
				if err := writeChartFile(ctx, destDir, f.Name, f.Data); err != nil {
					return nil, fmt.Errorf("error writing subchart file: %w", err)
				}
			}
		}
	}

	if wc.HelmChart.Schema != nil {
		schemaFile := filepath.Join(destDir, "values.schema.json")
		if err := writeChartFile(ctx, destDir, "values.schema.json", wc.HelmChart.Schema); err != nil {
			return nil, fmt.Errorf("error writing chart values schema: %w", err)
		}
		if err := ioutil.WriteFile(schemaFile, wc.HelmChart.Schema, os.ModePerm); err != nil {
			return nil, fmt.Errorf("unable to write %q: %w", schemaFile, err)
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

	return NewBundle(ctx, destDir, wc.HelmEnvSettings, wc.RegistryClient, wc.SecretsManager, BundleOptions{
		BuildChartDependenciesOpts:        wc.BuildChartDependenciesOpts,
		IgnoreInvalidAnnotationsAndLabels: wc.extraAnnotationsAndLabelsPostRenderer.IgnoreInvalidAnnotationsAndLabels,
		DisableDefaultValues:              wc.DisableDefaultValues,
	})
}

func writeChartFile(ctx context.Context, destDir, fileName string, fileData []byte) error {
	p := filepath.Join(destDir, fileName)
	dir := filepath.Dir(p)

	logboek.Context(ctx).Debug().LogF("Writing chart file %q\n", p)

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("error creating dir %q: %w", dir, err)
	}
	if err := ioutil.WriteFile(p, fileData, os.ModePerm); err != nil {
		return fmt.Errorf("unable to write %q: %w", p, err)
	}
	return nil
}
