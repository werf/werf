package secrets

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/deploy/secrets_manager"
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

func (secretsRuntimeData *SecretsRuntimeData) DecodeAndLoadSecrets(ctx context.Context, files []*chart.ChartExtenderBufferedFile, secretValueFiles []string, chartDir, secretsWorkingDir string, secretsManager *secrets_manager.SecretsManager) error {
	secretDirFiles := GetSecretDirFiles(files)
	secretValuesFiles := GetSecretValuesFiles(chartDir, files, SecretValuesFilesOptions{CustomFiles: secretValueFiles})

	var encoder *secret.YamlEncoder
	if len(secretDirFiles)+len(secretValuesFiles) > 0 {
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

	if len(secretValuesFiles) > 0 {
		if values, err := LoadChartSecretValueFiles(chartDir, secretValuesFiles, encoder); err != nil {
			return fmt.Errorf("error loading secret value files: %s", err)
		} else {
			secretsRuntimeData.DecodedSecretValues = values
			secretsRuntimeData.SecretValuesToMask = append(secretsRuntimeData.SecretValuesToMask, secretvalues.ExtractSecretValuesFromMap(values)...)
		}
	}

	return nil
}
