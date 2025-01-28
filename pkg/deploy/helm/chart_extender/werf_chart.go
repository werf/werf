package chart_extender

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/copystructure"
	"sigs.k8s.io/yaml"

	helm_v3 "github.com/werf/3p-helm/cmd/helm"
	"github.com/werf/3p-helm/pkg/chart"
	"github.com/werf/3p-helm/pkg/chart/loader"
	"github.com/werf/3p-helm/pkg/chartutil"
	"github.com/werf/3p-helm/pkg/cli"
	"github.com/werf/3p-helm/pkg/cli/values"
	"github.com/werf/3p-helm/pkg/getter"
	"github.com/werf/3p-helm/pkg/registry"
	"github.com/werf/3p-helm/pkg/werf/file"
	secrets2 "github.com/werf/3p-helm/pkg/werf/secrets"
	"github.com/werf/3p-helm/pkg/werf/secrets/runtimedata"
	"github.com/werf/common-go/pkg/secrets_manager"
	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/deploy/helm/chart_extender/helpers"
)

var _ chart.ChartExtender = (*WerfChart)(nil)

type WerfChartOptions struct {
	SecretValueFiles                  []string
	ExtraAnnotations                  map[string]string
	ExtraLabels                       map[string]string
	BuildChartDependenciesOpts        chart.BuildChartDependenciesOptions
	IgnoreInvalidAnnotationsAndLabels bool
	DisableDefaultValues              bool
	DisableDefaultSecretValues        bool
}

func NewWerfChart(
	ctx context.Context,
	chartFileReader file.ChartFileReader,
	chartDir string,
	projectDir string,
	helmEnvSettings *cli.EnvSettings,
	registryClient *registry.Client,
	opts WerfChartOptions,
) *WerfChart {
	wc := &WerfChart{
		ChartDir:         chartDir,
		ProjectDir:       projectDir,
		SecretValueFiles: opts.SecretValueFiles,
		HelmEnvSettings:  helmEnvSettings,
		RegistryClient:   registryClient,

		ChartFileReader: chartFileReader,

		ChartExtenderServiceValuesData: helpers.NewChartExtenderServiceValuesData(),

		DisableDefaultValues:       opts.DisableDefaultValues,
		DisableDefaultSecretValues: opts.DisableDefaultSecretValues,
		BuildChartDependenciesOpts: opts.BuildChartDependenciesOpts,
	}

	wc.AddExtraAnnotations(opts.ExtraAnnotations)
	wc.AddExtraLabels(opts.ExtraLabels)

	return wc
}

type WerfChart struct {
	HelmChart *chart.Chart

	ChartDir                   string
	ProjectDir                 string
	SecretValueFiles           []string
	HelmEnvSettings            *cli.EnvSettings
	RegistryClient             *registry.Client
	BuildChartDependenciesOpts chart.BuildChartDependenciesOptions
	DisableDefaultValues       bool
	DisableDefaultSecretValues bool

	ChartFileReader file.ChartFileReader

	extraAnnotations map[string]string
	extraLabels      map[string]string

	*helpers.ChartExtenderServiceValuesData
}

// SetHelmChart method for the chart.Extender interface
func (wc *WerfChart) SetHelmChart(c *chart.Chart) {
	wc.HelmChart = c
}

func (wc *WerfChart) MakeBundleValues(
	chrt *chart.Chart,
	inputVals map[string]interface{},
) (map[string]interface{}, error) {
	chartutil.DebugPrintValues(context.Background(), "input", inputVals)

	vals, err := chartutil.MergeInternal(context.Background(), inputVals, wc.GetServiceValues(), nil)
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

	chartutil.DebugPrintValues(context.Background(), "all", valsCopy)

	return valsCopy, nil
}

func (wc *WerfChart) MakeBundleSecretValues(
	ctx context.Context,
	secretsRuntimeData runtimedata.RuntimeData,
) (map[string]interface{}, error) {
	if chartutil.DebugSecretValues() {
		chartutil.DebugPrintValues(context.Background(), "secret", secretsRuntimeData.GetDecryptedSecretValues())
	}
	return secretsRuntimeData.GetEncodedSecretValues(ctx, secrets_manager.DefaultManager, wc.ProjectDir)
}

func (wc *WerfChart) AddExtraAnnotations(annotations map[string]string) {
	if wc.extraAnnotations == nil {
		wc.extraAnnotations = make(map[string]string)
	}

	for k, v := range annotations {
		wc.extraAnnotations[k] = v
	}
}

func (wc *WerfChart) AddExtraLabels(labels map[string]string) {
	if wc.extraLabels == nil {
		wc.extraLabels = make(map[string]string)
	}

	for k, v := range labels {
		wc.extraLabels[k] = v
	}
}

func (wc *WerfChart) GetExtraAnnotations() map[string]string {
	return wc.extraAnnotations
}

func (wc *WerfChart) GetExtraLabels() map[string]string {
	return wc.extraLabels
}

/*
 * CreateNewBundle creates new Bundle object with werf chart extensions taken into account.
 * inputVals could contain any custom values, which should be stored in the bundle.
 */
func (wc *WerfChart) CreateNewBundle(
	ctx context.Context,
	destDir, chartVersion string,
	vals *values.Options,
) (*Bundle, error) {
	chartPath := filepath.Join(wc.ProjectDir, wc.ChartDir)
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
	if chrt.SecretsRuntimeData != nil && !secrets_manager.DefaultManager.IsMissedSecretKeyModeEnabled() {
		vals, err := wc.MakeBundleSecretValues(ctx, chrt.SecretsRuntimeData)
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

	chartDirAbs := filepath.Join(wc.ProjectDir, wc.ChartDir)

	ignoreChartValuesFiles := []string{secrets2.DefaultSecretValuesFileName}

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

	if len(wc.extraAnnotations) > 0 {
		if err := writeBundleJsonMap(wc.extraAnnotations, filepath.Join(destDir, "extra_annotations.json")); err != nil {
			return nil, err
		}
	}

	if len(wc.extraLabels) > 0 {
		if err := writeBundleJsonMap(wc.extraLabels, filepath.Join(destDir, "extra_labels.json")); err != nil {
			return nil, err
		}
	}

	return NewBundle(ctx, destDir, BundleOptions{
		BuildChartDependenciesOpts: wc.BuildChartDependenciesOpts,
		DisableDefaultValues:       wc.DisableDefaultValues,
	})
}

func (wc *WerfChart) Type() string {
	return "chart"
}

func (wc *WerfChart) GetChartFileReader() file.ChartFileReader {
	return wc.ChartFileReader
}

func (wc *WerfChart) GetDisableDefaultValues() bool {
	return wc.DisableDefaultValues
}

func (wc *WerfChart) GetDisableDefaultSecretValues() bool {
	return wc.DisableDefaultSecretValues
}

func (wc *WerfChart) GetSecretValueFiles() []string {
	return wc.SecretValueFiles
}

func (wc *WerfChart) GetProjectDir() string {
	return wc.ProjectDir
}

func (wc *WerfChart) GetChartDir() string {
	return wc.ChartDir
}

func (wc *WerfChart) SetChartDir(dir string) {
	wc.ChartDir = dir
}

func (wc *WerfChart) GetBuildChartDependenciesOpts() chart.BuildChartDependenciesOptions {
	return wc.BuildChartDependenciesOpts
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
