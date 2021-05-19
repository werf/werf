package secrets

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/werf/werf/pkg/deploy/secrets_manager"
	"github.com/werf/werf/pkg/giterminism_manager"
	"github.com/werf/werf/pkg/secret"
	"github.com/werf/werf/pkg/util/secretvalues"
	"helm.sh/helm/v3/pkg/chart"
)

type SecretsRuntimeData struct {
	DecodedSecretValues    map[string]interface{}
	DecodedSecretFilesData map[string]string
	SecretValuesToMask     []string
}

func NewSecretsRuntimeData() *SecretsRuntimeData {
	return &SecretsRuntimeData{
		DecodedSecretFilesData: make(map[string]string),
	}
}

type DecodeAndLoadSecretsOptions struct {
	GiterminismManager      giterminism_manager.Interface
	CustomSecretValueFiles  []string
	LoadFromLocalFilesystem bool
}

func (secretsRuntimeData *SecretsRuntimeData) DecodeAndLoadSecrets(ctx context.Context, loadedChartFiles []*chart.ChartExtenderBufferedFile, chartDir, secretsWorkingDir string, secretsManager *secrets_manager.SecretsManager, opts DecodeAndLoadSecretsOptions) error {
	secretDirFiles := GetSecretDirFiles(loadedChartFiles)

	var loadedSecretValuesFiles []*chart.ChartExtenderBufferedFile
	if defaultSecretValues := GetDefaultSecretValuesFile(chartDir, loadedChartFiles); defaultSecretValues != nil {
		loadedSecretValuesFiles = append(loadedSecretValuesFiles, defaultSecretValues)
	}

	for _, customSecretValuesFileName := range opts.CustomSecretValueFiles {
		file := &chart.ChartExtenderBufferedFile{Name: customSecretValuesFileName}

		if opts.LoadFromLocalFilesystem {
			data, err := ioutil.ReadFile(customSecretValuesFileName)
			if err != nil {
				return fmt.Errorf("unable to read custom secret values file %q from local filesystem: %s", customSecretValuesFileName, err)
			}

			file.Data = data
		} else {
			data, err := opts.GiterminismManager.FileReader().ReadChartFile(ctx, customSecretValuesFileName)
			if err != nil {
				return fmt.Errorf("unable to read custom secret values file %q: %s", customSecretValuesFileName, err)
			}

			file.Data = data
		}

		loadedSecretValuesFiles = append(loadedSecretValuesFiles, file)
	}

	var encoder *secret.YamlEncoder
	if len(secretDirFiles)+len(loadedSecretValuesFiles) > 0 {
		if enc, err := secretsManager.GetYamlEncoder(ctx, secretsWorkingDir); err != nil {
			return err
		} else {
			encoder = enc
		}
	}

	if len(secretDirFiles) > 0 {
		if data, err := LoadChartSecretDirFilesData(chartDir, secretDirFiles, encoder); err != nil {
			return fmt.Errorf("error loading secret files data: %s", err)
		} else {
			secretsRuntimeData.DecodedSecretFilesData = data
			for _, fileData := range secretsRuntimeData.DecodedSecretFilesData {
				secretsRuntimeData.SecretValuesToMask = append(secretsRuntimeData.SecretValuesToMask, fileData)
			}
		}
	}

	if len(loadedSecretValuesFiles) > 0 {
		if values, err := LoadChartSecretValueFiles(chartDir, loadedSecretValuesFiles, encoder); err != nil {
			return fmt.Errorf("error loading secret value files: %s", err)
		} else {
			secretsRuntimeData.DecodedSecretValues = values
			secretsRuntimeData.SecretValuesToMask = append(secretsRuntimeData.SecretValuesToMask, secretvalues.ExtractSecretValuesFromMap(values)...)
		}
	}

	return nil
}
