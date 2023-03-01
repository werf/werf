package secrets

import (
	"context"
	"fmt"
	"io/ioutil"

	"helm.sh/helm/v3/pkg/chart"
	"sigs.k8s.io/yaml"

	"github.com/werf/werf/pkg/deploy/secrets_manager"
	"github.com/werf/werf/pkg/giterminism_manager"
	"github.com/werf/werf/pkg/secret"
	"github.com/werf/werf/pkg/util/secretvalues"
)

type SecretsRuntimeData struct {
	DecryptedSecretValues    map[string]interface{}
	DecryptedSecretFilesData map[string]string
	SecretValuesToMask       []string
}

func NewSecretsRuntimeData() *SecretsRuntimeData {
	return &SecretsRuntimeData{
		DecryptedSecretFilesData: make(map[string]string),
	}
}

type DecodeAndLoadSecretsOptions struct {
	GiterminismManager         giterminism_manager.Interface
	CustomSecretValueFiles     []string
	LoadFromLocalFilesystem    bool
	WithoutDefaultSecretValues bool
}

func (secretsRuntimeData *SecretsRuntimeData) DecodeAndLoadSecrets(ctx context.Context, loadedChartFiles []*chart.ChartExtenderBufferedFile, chartDir, secretsWorkingDir string, secretsManager *secrets_manager.SecretsManager, opts DecodeAndLoadSecretsOptions) error {
	secretDirFiles := GetSecretDirFiles(loadedChartFiles)

	var loadedSecretValuesFiles []*chart.ChartExtenderBufferedFile

	if !opts.WithoutDefaultSecretValues {
		if defaultSecretValues := GetDefaultSecretValuesFile(loadedChartFiles); defaultSecretValues != nil {
			loadedSecretValuesFiles = append(loadedSecretValuesFiles, defaultSecretValues)
		}
	}

	for _, customSecretValuesFileName := range opts.CustomSecretValueFiles {
		file := &chart.ChartExtenderBufferedFile{Name: customSecretValuesFileName}

		if opts.LoadFromLocalFilesystem {
			data, err := ioutil.ReadFile(customSecretValuesFileName)
			if err != nil {
				return fmt.Errorf("unable to read custom secret values file %q from local filesystem: %w", customSecretValuesFileName, err)
			}
			file.Data = data
		} else {
			data, err := opts.GiterminismManager.FileReader().ReadChartFile(ctx, customSecretValuesFileName)
			if err != nil {
				return fmt.Errorf("unable to read custom secret values file %q: %w", customSecretValuesFileName, err)
			}
			file.Data = data
		}

		loadedSecretValuesFiles = append(loadedSecretValuesFiles, file)
	}

	var encoder *secret.YamlEncoder
	if len(secretDirFiles)+len(loadedSecretValuesFiles) > 0 {
		if enc, err := secretsManager.GetYamlEncoder(ctx, secretsWorkingDir); err != nil {
			return fmt.Errorf("error getting secrets yaml encoder: %w", err)
		} else {
			encoder = enc
		}
	}

	if len(secretDirFiles) > 0 {
		if data, err := LoadChartSecretDirFilesData(chartDir, secretDirFiles, encoder); err != nil {
			return fmt.Errorf("error loading secret files data: %w", err)
		} else {
			secretsRuntimeData.DecryptedSecretFilesData = data
			for _, fileData := range secretsRuntimeData.DecryptedSecretFilesData {
				secretsRuntimeData.SecretValuesToMask = append(secretsRuntimeData.SecretValuesToMask, fileData)
			}
		}
	}

	if len(loadedSecretValuesFiles) > 0 {
		if values, err := LoadChartSecretValueFiles(chartDir, loadedSecretValuesFiles, encoder); err != nil {
			return fmt.Errorf("error loading secret value files: %w", err)
		} else {
			secretsRuntimeData.DecryptedSecretValues = values
			secretsRuntimeData.SecretValuesToMask = append(secretsRuntimeData.SecretValuesToMask, secretvalues.ExtractSecretValuesFromMap(values)...)
		}
	}

	return nil
}

func (secretsRuntimeData *SecretsRuntimeData) GetEncodedSecretValues(ctx context.Context, secretsManager *secrets_manager.SecretsManager, secretsWorkingDir string) (map[string]interface{}, error) {
	if len(secretsRuntimeData.DecryptedSecretValues) == 0 {
		return nil, nil
	}

	// FIXME: secrets encoder should receive interface{} raw data instead of []byte yaml data

	var encoder *secret.YamlEncoder
	if enc, err := secretsManager.GetYamlEncoder(ctx, secretsWorkingDir); err != nil {
		return nil, fmt.Errorf("error getting secrets yaml encoder: %w", err)
	} else {
		encoder = enc
	}

	decryptedSecretsData, err := yaml.Marshal(secretsRuntimeData.DecryptedSecretValues)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal decrypted secrets yaml: %w", err)
	}

	encryptedSecretsData, err := encoder.EncryptYamlData(decryptedSecretsData)
	if err != nil {
		return nil, fmt.Errorf("unable to encrypt secrets data: %w", err)
	}

	var encryptedData map[string]interface{}
	if err := yaml.Unmarshal(encryptedSecretsData, &encryptedData); err != nil {
		return nil, fmt.Errorf("unable to unmarshal encrypted secrets data: %w", err)
	}

	return encryptedData, nil
}
