package werf_chart

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/deploy/secret"
	"github.com/werf/werf/pkg/deploy_v2/helm_v3"
	"github.com/werf/werf/pkg/deploy_v2/lock_manager"
	"github.com/werf/werf/pkg/util/secretvalues"
	"github.com/werf/werf/pkg/werf"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
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

func NewWerfChart(opts WerfChartOptions) *WerfChart {
	wc := &WerfChart{
		ReleaseName: opts.ReleaseName,
		ChartDir:    opts.ChartDir,

		SecretValueFiles: opts.SecretValueFiles,
		ExtraAnnotationsAndLabelsPostRenderer: helm_v3.NewExtraAnnotationsAndLabelsPostRenderer(
			map[string]string{"werf.io/version": werf.Version},
			nil,
		),

		LockManager:    opts.LockManager,
		SecretsManager: opts.SecretsManager,

		decodedSecretFilesData: make(map[string]string, 0),
	}

	wc.ExtraAnnotationsAndLabelsPostRenderer.Add(opts.ExtraAnnotations, opts.ExtraLabels)

	return wc
}

type WerfChart struct {
	HelmChart *chart.Chart

	ReleaseName      string
	ChartDir         string
	SecretValueFiles []string

	ExtraAnnotationsAndLabelsPostRenderer *helm_v3.ExtraAnnotationsAndLabelsPostRenderer
	LockManager                           *lock_manager.LockManager
	SecretsManager                        secret.Manager

	chartMetadataFromWerfConfig *chart.Metadata
	decodedSecretValues         map[string]interface{}
	decodedSecretFilesData      map[string]string
	secretValuesToMask          []string
}

func (wc *WerfChart) SetupChart(c *chart.Chart) error {
	wc.HelmChart = c
	return nil
}

func (wc *WerfChart) AfterLoad() error {
	secretValuesFiles := []string{}
	defaultSecretValuesFile := filepath.Join(wc.ChartDir, DefaultSecretValuesFileName)
	if _, err := os.Stat(defaultSecretValuesFile); !os.IsNotExist(err) {
		secretValuesFiles = append(secretValuesFiles, defaultSecretValuesFile)
	}
	for _, path := range wc.SecretValueFiles {
		secretValuesFiles = append(secretValuesFiles, path)
	}
	for _, path := range secretValuesFiles {
		if decodedValues, err := DecodeSecretValuesFile(path, wc.SecretsManager); err != nil {
			return fmt.Errorf("unable to decode secret values file %q: %s", path, err)
		} else {
			wc.decodedSecretValues = chartutil.CoalesceTables(decodedValues, wc.decodedSecretValues)
			wc.secretValuesToMask = append(wc.secretValuesToMask, secretvalues.ExtractSecretValuesFromMap(decodedValues)...)
		}
	}

	secretDir := filepath.Join(wc.ChartDir, SecretDirName)
	if _, err := os.Stat(secretDir); !os.IsNotExist(err) {
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

	if wc.HelmChart.Metadata == nil && wc.chartMetadataFromWerfConfig != nil {
		wc.HelmChart.Metadata = wc.chartMetadataFromWerfConfig
	}

	return nil
}

func (wc *WerfChart) MakeValues(inputVals map[string]interface{}) (map[string]interface{}, error) {
	vals := make(map[string]interface{})
	chartutil.CoalesceTables(vals, wc.decodedSecretValues)
	chartutil.CoalesceTables(vals, inputVals)
	return vals, nil
}

func (wc *WerfChart) SetWerfConfig(werfConfig *config.WerfConfig) error {
	wc.ExtraAnnotationsAndLabelsPostRenderer.Add(map[string]string{
		"project.werf.io/name": werfConfig.Meta.Project,
	}, nil)

	wc.chartMetadataFromWerfConfig = &chart.Metadata{
		APIVersion: "v1",
		Name:       werfConfig.Meta.Project,
		Version:    "1.0.0",
	}

	return nil
}

func (wc *WerfChart) SetEnv(env string) error {
	wc.ExtraAnnotationsAndLabelsPostRenderer.Add(map[string]string{
		"project.werf.io/env": env,
	}, nil)

	return nil
}

func (wc *WerfChart) WrapInstall(ctx context.Context, installFunc func() error) error {
	return wc.lockReleaseWrapper(ctx, installFunc)
}

func (wc *WerfChart) WrapUpgrade(ctx context.Context, upgradeFunc func() error) error {
	return wc.lockReleaseWrapper(ctx, upgradeFunc)
}

func (wc *WerfChart) lockReleaseWrapper(ctx context.Context, commandFunc func() error) error {
	if lock, err := wc.LockManager.LockRelease(ctx, wc.ReleaseName); err != nil {
		return err
	} else {
		defer wc.LockManager.Unlock(lock)
	}
	return commandFunc()
}
