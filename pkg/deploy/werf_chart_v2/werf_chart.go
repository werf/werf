package werf_chart_v2

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/werf/werf/pkg/deploy/lock_manager"

	"github.com/werf/werf/pkg/util/secretvalues"

	"github.com/werf/werf/pkg/config"

	"github.com/werf/werf/pkg/deploy/helm_v3"
	"github.com/werf/werf/pkg/deploy/secret"
	"github.com/werf/werf/pkg/werf"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli/values"
)

const (
	DefaultSecretValuesFileName = "secret-values.yaml"
	SecretDirName               = "secret"
)

func NewWerfChart() *WerfChart {
	return &WerfChart{
		ExtraAnnotationsAndLabelsPostRenderer: helm_v3.NewExtraAnnotationsAndLabelsPostRenderer(
			map[string]string{"werf.io/version": werf.Version},
			nil,
		),
		ValueOpts:              &values.Options{},
		decodedSecretFilesData: make(map[string]string, 0),
	}
}

type WerfChart struct {
	ReleaseName                           string
	ChartDir                              string
	ChartConfig                           *chart.Metadata
	SecretValues                          []map[string]interface{}
	ExtraAnnotationsAndLabelsPostRenderer *helm_v3.ExtraAnnotationsAndLabelsPostRenderer
	ValueOpts                             *values.Options

	LockManager    *lock_manager.LockManager
	SecretsManager secret.Manager

	decodedSecretFilesData map[string]string
	secretValuesToMask     []string
	initialized            bool
}

type WerfChartInitOptions struct {
	LockManager    *lock_manager.LockManager
	SecretsManager secret.Manager

	ReleaseName       string
	ChartDir          string
	SecretValuesFiles []string
	ExtraAnnotations  map[string]string
	ExtraLabels       map[string]string
}

// Load secrets, validate, etc.
func (wc *WerfChart) Init(opts WerfChartInitOptions) error {
	if wc.initialized {
		panic(fmt.Sprintf("werf chart %#v already initialized", *wc))
	}

	wc.LockManager = opts.LockManager
	wc.SecretsManager = opts.SecretsManager
	wc.ReleaseName = opts.ReleaseName
	wc.ChartDir = opts.ChartDir

	secretValuesFiles := []string{}
	defaultSecretValuesFile := filepath.Join(wc.ChartDir, DefaultSecretValuesFileName)
	if _, err := os.Stat(defaultSecretValuesFile); !os.IsNotExist(err) {
		secretValuesFiles = append(secretValuesFiles, defaultSecretValuesFile)
	}
	for _, path := range opts.SecretValuesFiles {
		secretValuesFiles = append(secretValuesFiles, path)
	}
	for _, path := range secretValuesFiles {
		if decodedValues, err := DecodeSecretValuesFile(path, wc.SecretsManager); err != nil {
			return fmt.Errorf("unable to decode secret values file %q: %s", path, err)
		} else {
			wc.ValueOpts.RawValues = append(wc.ValueOpts.RawValues, decodedValues)
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

	wc.ExtraAnnotationsAndLabelsPostRenderer.Add(opts.ExtraAnnotations, opts.ExtraLabels)

	wc.initialized = true
	return nil
}

func (wc *WerfChart) SetWerfConfig(werfConfig *config.WerfConfig) error {
	wc.ExtraAnnotationsAndLabelsPostRenderer.Add(map[string]string{
		"project.werf.io/name": werfConfig.Meta.Project,
	}, nil)

	wc.ChartConfig = &chart.Metadata{
		APIVersion: "v1",
		Name:       werfConfig.Meta.Project,
		Version:    "0.1.0", // FIXME
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
