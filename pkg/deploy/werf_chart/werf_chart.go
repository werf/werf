package werf_chart

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

	"helm.sh/helm/v3/pkg/postrender"

	"github.com/werf/logboek"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"

	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/deploy/helm"
	"github.com/werf/werf/pkg/deploy/lock_manager"
	"github.com/werf/werf/pkg/deploy/secret"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/giterminism_inspector"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/util/secretvalues"
)

const (
	DefaultSecretValuesFileName = "secret-values.yaml"
	SecretDirName               = "secret"
)

type WerfChartOptions struct {
	ReleaseName string
	ChartDir    string

	SecretValueFiles []string
	ExtraAnnotations map[string]string
	ExtraLabels      map[string]string

	LockManager    *lock_manager.LockManager
	SecretsManager secret.Manager
}

func NewWerfChart(ctx context.Context, localGitRepo *git_repo.Local, projectDir string, opts WerfChartOptions) *WerfChart {
	wc := &WerfChart{
		Ctx: ctx,

		ReleaseName: opts.ReleaseName,
		ChartDir:    opts.ChartDir,
		ProjectDir:  projectDir,

		SecretValueFiles:                      opts.SecretValueFiles,
		ExtraAnnotationsAndLabelsPostRenderer: helm.NewExtraAnnotationsAndLabelsPostRenderer(nil, nil),

		LockManager:    opts.LockManager,
		SecretsManager: opts.SecretsManager,
		LocalGitRepo:   localGitRepo,

		decodedSecretFilesData: make(map[string]string, 0),
	}

	wc.ExtraAnnotationsAndLabelsPostRenderer.Add(opts.ExtraAnnotations, opts.ExtraLabels)

	return wc
}

type WerfChart struct {
	Ctx       context.Context
	HelmChart *chart.Chart

	ReleaseName      string
	ChartDir         string
	ProjectDir       string
	SecretValueFiles []string

	ExtraAnnotationsAndLabelsPostRenderer *helm.ExtraAnnotationsAndLabelsPostRenderer
	LockManager                           *lock_manager.LockManager
	SecretsManager                        secret.Manager
	LocalGitRepo                          *git_repo.Local

	werfConfig             *config.WerfConfig
	decodedSecretValues    map[string]interface{}
	decodedSecretFilesData map[string]string
	secretValuesToMask     []string
	serviceValues          map[string]interface{}
}

func (wc *WerfChart) GetPostRenderer() (postrender.PostRenderer, error) {
	return wc.ExtraAnnotationsAndLabelsPostRenderer, nil
}

func (wc *WerfChart) SetupChart(c *chart.Chart) error {
	wc.HelmChart = c
	return nil
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
		if decodedValues, err := DecodeSecretValuesFileFromFilesystem(wc.Ctx, path, wc.SecretsManager); err != nil {
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

	commit, err := wc.LocalGitRepo.HeadCommit(wc.Ctx)
	if err != nil {
		return fmt.Errorf("unable to get local repo head commit: %s", err)
	}

	var chartDir string
	if isSymlink, linkDest, err := wc.LocalGitRepo.CheckAndReadCommitSymlink(wc.Ctx, wc.ChartDir, commit); err != nil {
		return fmt.Errorf("error checking %q is symlink in the local git repo commit %s: %s", wc.ChartDir, commit, err)
	} else if isSymlink {
		chartDir = string(linkDest)
	} else {
		chartDir = wc.ChartDir
	}

	defaultSecretValuesFile := filepath.Join(chartDir, DefaultSecretValuesFileName)
	if exists, err := wc.LocalGitRepo.IsCommitFileExists(wc.Ctx, commit, defaultSecretValuesFile); err != nil {
		return fmt.Errorf("error checking existence of the file %q in the local git repo commit %s: %s", defaultSecretValuesFile, commit, err)
	} else if exists {
		logboek.Context(wc.Ctx).Debug().LogF("Check %s exists in the local git repo commit %s: FOUND\n", defaultSecretValuesFile, commit)
		secretValuesFiles = append(secretValuesFiles, defaultSecretValuesFile)
	} else {
		logboek.Context(wc.Ctx).Debug().LogF("Check %s exists in the local git repo commit %s: NOT FOUND\n", defaultSecretValuesFile, commit)
	}

	for _, path := range wc.SecretValueFiles {
		secretValuesFiles = append(secretValuesFiles, path)
	}

	for _, path := range secretValuesFiles {
		logboek.Context(wc.Ctx).Debug().LogF("Decoding secret values file %q\n", path)

		var decodedValues map[string]interface{}

		commit, err := wc.LocalGitRepo.HeadCommit(wc.Ctx)
		if err != nil {
			return fmt.Errorf("unable to get local repo head commit: %s", err)
		}

		if vals, err := DecodeSecretValuesFileFromGitCommit(wc.Ctx, path, commit, wc.LocalGitRepo, wc.SecretsManager, wc.ProjectDir); err != nil {
			return fmt.Errorf("unable to decode secret values file %q: %s", path, err)
		} else {
			decodedValues = vals
		}

		wc.decodedSecretValues = chartutil.CoalesceTables(decodedValues, wc.decodedSecretValues)
		wc.secretValuesToMask = append(wc.secretValuesToMask, secretvalues.ExtractSecretValuesFromMap(decodedValues)...)
	}

	secretDir := filepath.Join(wc.ChartDir, SecretDirName)
	if exists, err := wc.LocalGitRepo.IsCommitDirectoryExists(wc.Ctx, secretDir, commit); err != nil {
		return fmt.Errorf("error checking existence of directory %s in the local git repo commit %s: %s", secretDir, commit, err)
	} else if exists {
		var secretFilesToDecode []string

		if paths, err := wc.LocalGitRepo.GetCommitFilePathList(wc.Ctx, commit); err != nil {
			return fmt.Errorf("error getting file path list for the local git repo commit %s: %s", commit, err)
		} else {
			for _, path := range paths {
				if util.IsSubpathOfBasePath(secretDir, path) {
					secretFilesToDecode = append(secretFilesToDecode, path)
				}
			}
		}

		for _, path := range secretFilesToDecode {
			data, err := wc.LocalGitRepo.ReadCommitFile(wc.Ctx, commit, path)
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

func (wc *WerfChart) AfterLoad() error {
	if giterminism_inspector.LooseGiterminism || wc.LocalGitRepo == nil {
		if err := wc.loadSecretsFromFilesystem(); err != nil {
			return err
		}
	} else {
		if err := wc.loadSecretsFromLocalGitRepo(); err != nil {
			return err
		}
	}

	if wc.HelmChart.Metadata == nil {
		wc.HelmChart.Metadata = &chart.Metadata{
			APIVersion: chart.APIVersionV2,
		}
	}

	if wc.HelmChart.Metadata.Name == "" {
		wc.HelmChart.Metadata.Name = "stub_name"
	}

	if wc.werfConfig != nil {
		wc.HelmChart.Metadata.Name = wc.werfConfig.Meta.Project
	}

	if wc.HelmChart.Metadata.Version == "" {
		wc.HelmChart.Metadata.Version = "1.0.0"
	}

	wc.HelmChart.Templates = append(wc.HelmChart.Templates, &chart.File{
		Name: "templates/_werf_helpers.tpl",
		Data: []byte(TemplateHelpers),
	})

	return nil
}

func (wc *WerfChart) MakeValues(inputVals map[string]interface{}) (map[string]interface{}, error) {
	vals := make(map[string]interface{})
	chartutil.CoalesceTables(vals, wc.serviceValues) // NOTE: service values will not be saved into the marshalled release
	chartutil.CoalesceTables(vals, wc.decodedSecretValues)
	chartutil.CoalesceTables(vals, inputVals)
	return vals, nil
}

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

func (wc *WerfChart) SetWerfConfig(werfConfig *config.WerfConfig) error {
	wc.ExtraAnnotationsAndLabelsPostRenderer.Add(map[string]string{
		"project.werf.io/name": werfConfig.Meta.Project,
	}, nil)

	wc.werfConfig = werfConfig

	return nil
}

func (wc *WerfChart) SetEnv(env string) error {
	wc.ExtraAnnotationsAndLabelsPostRenderer.Add(map[string]string{
		"project.werf.io/env": env,
	}, nil)

	return nil
}

func (wc *WerfChart) SetServiceValues(vals map[string]interface{}) error {
	wc.serviceValues = vals
	return nil
}

func (wc *WerfChart) WrapTemplate(ctx context.Context, templateFunc func() error) error {
	return templateFunc()
}

func (wc *WerfChart) WrapInstall(ctx context.Context, installFunc func() error) error {
	return wc.lockReleaseWrapper(ctx, installFunc)
}

func (wc *WerfChart) WrapUpgrade(ctx context.Context, upgradeFunc func() error) error {
	return wc.lockReleaseWrapper(ctx, upgradeFunc)
}

func (wc *WerfChart) WrapUninstall(ctx context.Context, uninstallFunc func() error, withNamespace bool) error {
	if withNamespace {
		// FIXME: Maybe store deploy locks in the namespace object itself.
		// FIXME: The problem is: werf cannot delete namespace while it holds a lock for current release, which stored in this same namespace.
		return uninstallFunc()
	}

	return wc.lockReleaseWrapper(ctx, uninstallFunc)
}

func (wc *WerfChart) lockReleaseWrapper(ctx context.Context, commandFunc func() error) error {
	if wc.LockManager != nil {
		if lock, err := wc.LockManager.LockRelease(ctx, wc.ReleaseName); err != nil {
			return err
		} else {
			defer wc.LockManager.Unlock(lock)
		}
	}
	return commandFunc()
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

	if wc.ExtraAnnotationsAndLabelsPostRenderer.ExtraAnnotations != nil {
		if err := writeBundleJsonMap(wc.ExtraAnnotationsAndLabelsPostRenderer.ExtraAnnotations, filepath.Join(destDir, "extra_annotations.json")); err != nil {
			return nil, err
		}
	}

	if wc.ExtraAnnotationsAndLabelsPostRenderer.ExtraLabels != nil {
		if err := writeBundleJsonMap(wc.ExtraAnnotationsAndLabelsPostRenderer.ExtraLabels, filepath.Join(destDir, "extra_labels.json")); err != nil {
			return nil, err
		}
	}

	return NewBundle(destDir), nil
}
