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
	"unicode"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/util/secretvalues"

	"github.com/werf/werf/pkg/deploy/helm/command_helpers"
	"github.com/werf/werf/pkg/util"
	"helm.sh/helm/v3/pkg/cli"

	"github.com/werf/werf/pkg/giterminism"

	"helm.sh/helm/v3/pkg/postrender"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"

	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/deploy/helm"
	"github.com/werf/werf/pkg/deploy/secret"
	"github.com/werf/werf/pkg/giterminism_inspector"
)

const (
	DefaultSecretValuesFileName = "secret-values.yaml"
	SecretDirName               = "secret"
)

type WerfChartOptions struct {
	SecretValueFiles           []string
	ExtraAnnotations           map[string]string
	ExtraLabels                map[string]string
	BuildChartDependenciesOpts command_helpers.BuildChartDependenciesOptions
}

func NewWerfChart(giterminismManager giterminism.Manager, secretManager secret.Manager, projectDir, chartDir string, helmEnvSettings *cli.EnvSettings, opts WerfChartOptions) *WerfChart {
	wc := &WerfChart{
		ProjectDir:       projectDir,
		ChartDir:         chartDir,
		SecretValueFiles: opts.SecretValueFiles,
		HelmEnvSettings:  helmEnvSettings,

		GiterminismManager: giterminismManager,
		SecretsManager:     secretManager,

		extraAnnotationsAndLabelsPostRenderer: helm.NewExtraAnnotationsAndLabelsPostRenderer(nil, nil),
		decodedSecretFilesData:                make(map[string]string, 0),
	}

	wc.extraAnnotationsAndLabelsPostRenderer.Add(opts.ExtraAnnotations, opts.ExtraLabels)

	return wc
}

type WerfChart struct {
	HelmChart *chart.Chart

	ProjectDir                 string
	ChartDir                   string
	SecretValueFiles           []string
	HelmEnvSettings            *cli.EnvSettings
	BuildChartDependenciesOpts command_helpers.BuildChartDependenciesOptions

	SecretsManager     secret.Manager
	GiterminismManager giterminism.Manager

	extraAnnotationsAndLabelsPostRenderer *helm.ExtraAnnotationsAndLabelsPostRenderer
	werfConfig                            *config.WerfConfig
	decodedSecretValues                   map[string]interface{}
	decodedSecretFilesData                map[string]string
	secretValuesToMask                    []string
	serviceValues                         map[string]interface{}
	chartExtenderContext                  context.Context
}

// ChartCreated method for the chart.Extender interface
func (wc *WerfChart) ChartCreated(c *chart.Chart) error {
	wc.HelmChart = c
	return nil
}

// ChartLoaded method for the chart.Extender interface
func (wc *WerfChart) ChartLoaded(files []*chart.ChartExtenderBufferedFile) error {
	// TODO: Remove loose giterminism, load secrets from provided buffered-files param, do not read any files by itself
	if wc.SecretsManager != nil {
		if giterminism_inspector.LooseGiterminism {
			if err := wc.loadSecretsFromFilesystem(); err != nil {
				return err
			}
		} else {
			if err := wc.loadSecretsFromLocalGitRepo(); err != nil {
				return err
			}
		}
	}

	var opts GetHelmChartMetadataOptions
	if wc.werfConfig != nil {
		opts.OverrideName = wc.werfConfig.Meta.Project
	}
	opts.DefaultVersion = "1.0.0"
	wc.HelmChart.Metadata = GetHelmChartMetadataWithOverrides(wc.HelmChart.Metadata, opts)

	wc.HelmChart.Templates = append(wc.HelmChart.Templates, &chart.File{
		Name: "templates/_werf_helpers.tpl",
		Data: []byte(TemplateHelpers),
	})

	return nil
}

// ChartDependenciesLoaded method for the chart.Extender interface
func (wc *WerfChart) ChartDependenciesLoaded() error {
	return nil
}

// MakeValues method for the chart.Extender interface
func (wc *WerfChart) MakeValues(inputVals map[string]interface{}) (map[string]interface{}, error) {
	vals := make(map[string]interface{})
	chartutil.CoalesceTables(vals, wc.serviceValues) // NOTE: service values will not be saved into the marshalled release
	chartutil.CoalesceTables(vals, wc.decodedSecretValues)
	chartutil.CoalesceTables(vals, inputVals)
	return vals, nil
}

// SetupTemplateFuncs method for the chart.Extender interface
func (wc *WerfChart) SetupTemplateFuncs(t *template.Template, funcMap template.FuncMap) {
	funcMap["werf_secret_file"] = func(secretRelativePath string) (string, error) {
		if path.IsAbs(secretRelativePath) {
			return "", fmt.Errorf("expected relative secret file path, given path %v", secretRelativePath)
		}

		decodedData, ok := wc.decodedSecretFilesData[secretRelativePath]

		if !ok {
			var secretFiles []string
			for key := range wc.decodedSecretFilesData {
				secretFiles = append(secretFiles, key)
			}

			return "", fmt.Errorf("secret file '%s' not found, you may use one of the following: '%s'", secretRelativePath, strings.Join(secretFiles, "', '"))
		}

		return decodedData, nil
	}

	SetupIncludeWrapperFuncs(funcMap)
}

// LoadDir method for the chart.Extender interface
func (wc *WerfChart) LoadDir(dir string) (bool, []*chart.ChartExtenderBufferedFile, error) {
	// TODO: Remove loose giterminism, return always true
	if giterminism_inspector.LooseGiterminism {
		return false, nil, nil
	}

	res, err := GiterministicFilesLoader(wc.chartExtenderContext, wc.GiterminismManager, dir, wc.HelmEnvSettings, wc.BuildChartDependenciesOpts)
	return true, res, err
}

// LocateChart method for the chart.Extender interface
func (wc *WerfChart) LocateChart(name string, settings *cli.EnvSettings) (bool, string, error) {
	// TODO: Remove loose giterminism, return always true
	if giterminism_inspector.LooseGiterminism {
		return false, "", nil
	}

	res, err := wc.GiterminismManager.FileReader().LocateChart(wc.chartExtenderContext, name, settings)
	return true, res, err
}

// ReadFile method for the chart.Extender interface
func (wc *WerfChart) ReadFile(filePath string) (bool, []byte, error) {
	// TODO: Remove loose giterminism, return always true
	if giterminism_inspector.LooseGiterminism {
		return false, nil, nil
	}

	res, err := wc.GiterminismManager.FileReader().ReadChartFile(wc.chartExtenderContext, filePath)
	return true, res, err
}

func (wc *WerfChart) GetPostRenderer() (postrender.PostRenderer, error) {
	return wc.extraAnnotationsAndLabelsPostRenderer, nil
}

func (wc *WerfChart) SetChartExtenderContext(ctx context.Context) {
	wc.chartExtenderContext = ctx
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

func (wc *WerfChart) SetServiceValues(vals map[string]interface{}) error {
	wc.serviceValues = vals
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

	if vals, err := wc.MakeValues(inputVals); err != nil {
		return nil, fmt.Errorf("unable to coalesce input values: %s", err)
	} else if valsData, err := json.Marshal(vals); err != nil {
		return nil, fmt.Errorf("unable to prepare values: %s", err)
	} else {
		valuesFile := filepath.Join(destDir, "values.yaml")
		if err := ioutil.WriteFile(valuesFile, append(valsData, []byte("\n")...), os.ModePerm); err != nil {
			return nil, fmt.Errorf("unable to write %q: %s", valuesFile, err)
		}
	}

	if wc.HelmChart.Metadata != nil {
		bundleMetadata := *wc.HelmChart.Metadata
		// Force api v2
		bundleMetadata.APIVersion = chart.APIVersionV2

		chartYamlFile := filepath.Join(destDir, "Chart.yaml")
		if data, err := json.Marshal(bundleMetadata); err != nil {
			return nil, fmt.Errorf("unable to prepare Chart.yaml data: %s", err)
		} else if err := ioutil.WriteFile(chartYamlFile, append(data, []byte("\n")...), os.ModePerm); err != nil {
			return nil, fmt.Errorf("unable to write %q: %s", chartYamlFile, err)
		}
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

	return NewBundle(destDir), nil
}

func (wc *WerfChart) loadSecretsFromFilesystem() error {
	secretValuesFiles := []string{}
	defaultSecretValuesFile := filepath.Join(wc.ChartDir, DefaultSecretValuesFileName)
	if exists, err := util.RegularFileExists(defaultSecretValuesFile); err != nil {
		return fmt.Errorf("unable to check file %s existence: %s", defaultSecretValuesFile, err)
	} else if exists {
		secretValuesFiles = append(secretValuesFiles, defaultSecretValuesFile)
	}
	for _, path := range wc.SecretValueFiles {
		secretValuesFiles = append(secretValuesFiles, path)
	}
	for _, path := range secretValuesFiles {
		if decodedValues, err := DecodeSecretValuesFileFromFilesystem(wc.chartExtenderContext, path, wc.SecretsManager); err != nil {
			return fmt.Errorf("unable to decode secret values file %q: %s", path, err)
		} else {
			wc.decodedSecretValues = chartutil.CoalesceTables(decodedValues, wc.decodedSecretValues)
			wc.secretValuesToMask = append(wc.secretValuesToMask, secretvalues.ExtractSecretValuesFromMap(decodedValues)...)
		}
	}

	secretDir := filepath.Join(wc.ChartDir, SecretDirName)
	if exists, err := util.DirExists(secretDir); err != nil {
		return fmt.Errorf("unable to check dir %s existence: %s", secretDir, err)
	} else if exists {
		if err := filepath.Walk(secretDir, func(path string, info os.FileInfo, accessErr error) error {
			if accessErr != nil {
				return fmt.Errorf("error accessing file %s: %s", path, accessErr)
			}

			if info.Mode().IsDir() {
				return nil
			}

			data, err := ioutil.ReadFile(path)
			if err != nil {
				return fmt.Errorf("error reading file %s: %s", path, err)
			}

			decodedData, err := wc.SecretsManager.Decrypt([]byte(strings.TrimRightFunc(string(data), unicode.IsSpace)))
			if err != nil {
				return fmt.Errorf("error decoding %s: %s", path, err)
			}

			relativePath, err := filepath.Rel(secretDir, path)
			if err != nil {
				panic(err)
			}

			wc.decodedSecretFilesData[filepath.ToSlash(relativePath)] = string(decodedData)
			wc.secretValuesToMask = append(wc.secretValuesToMask, string(decodedData))

			return nil
		}); err != nil {
			return fmt.Errorf("unable to read secrets from %s directory: %s", secretDir, err)
		}
	}

	return nil
}

func (wc *WerfChart) loadSecretsFromLocalGitRepo() error {
	var secretValuesFiles []string

	commit, err := wc.GiterminismManager.LocalGitRepo().HeadCommit(wc.chartExtenderContext)
	if err != nil {
		return fmt.Errorf("unable to get local repo head commit: %s", err)
	}

	var chartDir string
	if isSymlink, linkDest, err := wc.GiterminismManager.LocalGitRepo().CheckAndReadCommitSymlink(wc.chartExtenderContext, wc.ChartDir, commit); err != nil {
		return fmt.Errorf("error checking %q is symlink in the local git repo commit %s: %s", wc.ChartDir, commit, err)
	} else if isSymlink {
		chartDir = string(linkDest)
	} else {
		chartDir = wc.ChartDir
	}

	defaultSecretValuesFile := filepath.Join(chartDir, DefaultSecretValuesFileName)
	if exists, err := wc.GiterminismManager.LocalGitRepo().IsCommitFileExists(wc.chartExtenderContext, commit, defaultSecretValuesFile); err != nil {
		return fmt.Errorf("error checking existence of the file %q in the local git repo commit %s: %s", defaultSecretValuesFile, commit, err)
	} else if exists {
		logboek.Context(wc.chartExtenderContext).Debug().LogF("Check %s exists in the local git repo commit %s: FOUND\n", defaultSecretValuesFile, commit)
		secretValuesFiles = append(secretValuesFiles, defaultSecretValuesFile)
	} else {
		logboek.Context(wc.chartExtenderContext).Debug().LogF("Check %s exists in the local git repo commit %s: NOT FOUND\n", defaultSecretValuesFile, commit)
	}

	for _, path := range wc.SecretValueFiles {
		secretValuesFiles = append(secretValuesFiles, path)
	}

	for _, path := range secretValuesFiles {
		logboek.Context(wc.chartExtenderContext).Debug().LogF("Decoding secret values file %q\n", path)

		var decodedValues map[string]interface{}

		commit, err := wc.GiterminismManager.LocalGitRepo().HeadCommit(wc.chartExtenderContext)
		if err != nil {
			return fmt.Errorf("unable to get local repo head commit: %s", err)
		}

		if vals, err := DecodeSecretValuesFileFromGitCommit(wc.chartExtenderContext, path, commit, wc.GiterminismManager.LocalGitRepo(), wc.SecretsManager, wc.ProjectDir); err != nil {
			return fmt.Errorf("unable to decode secret values file %q: %s", path, err)
		} else {
			decodedValues = vals
		}

		wc.decodedSecretValues = chartutil.CoalesceTables(decodedValues, wc.decodedSecretValues)
		wc.secretValuesToMask = append(wc.secretValuesToMask, secretvalues.ExtractSecretValuesFromMap(decodedValues)...)
	}

	secretDir := filepath.Join(wc.ChartDir, SecretDirName)
	if exists, err := wc.GiterminismManager.LocalGitRepo().IsCommitDirectoryExists(wc.chartExtenderContext, secretDir, commit); err != nil {
		return fmt.Errorf("error checking existence of directory %s in the local git repo commit %s: %s", secretDir, commit, err)
	} else if exists {
		var secretFilesToDecode []string

		if paths, err := wc.GiterminismManager.LocalGitRepo().GetCommitFilePathList(wc.chartExtenderContext, commit); err != nil {
			return fmt.Errorf("error getting file path list for the local git repo commit %s: %s", commit, err)
		} else {
			for _, path := range paths {
				if util.IsSubpathOfBasePath(secretDir, path) {
					secretFilesToDecode = append(secretFilesToDecode, path)
				}
			}
		}

		for _, path := range secretFilesToDecode {
			data, err := wc.GiterminismManager.LocalGitRepo().ReadCommitFile(wc.chartExtenderContext, commit, path)
			if err != nil {
				return fmt.Errorf("error reading file %s from the local git repo commit %s: %s", path, commit, err)
			}

			decodedData, err := wc.SecretsManager.Decrypt([]byte(strings.TrimRightFunc(string(data), unicode.IsSpace)))
			if err != nil {
				return fmt.Errorf("error decoding %s: %s", path, err)
			}

			relativePath, err := filepath.Rel(secretDir, path)
			if err != nil {
				panic(err)
			}

			wc.decodedSecretFilesData[filepath.ToSlash(relativePath)] = string(decodedData)
			wc.secretValuesToMask = append(wc.secretValuesToMask, string(decodedData))
		}
	}

	return nil
}
