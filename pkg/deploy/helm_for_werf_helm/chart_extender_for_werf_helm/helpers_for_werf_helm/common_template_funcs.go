package helpers_for_werf_helm

import (
	"fmt"
	"path"
	"strings"
	"text/template"

	secrets "github.com/werf/werf/v2/pkg/deploy/helm_for_werf_helm/chart_extender_for_werf_helm/helpers_for_werf_helm/secrets_for_werf_helm"
)

func SetupIncludeWrapperFuncs(funcMap template.FuncMap) {
	helmIncludeFunc := funcMap["include"].(func(name string, data interface{}) (string, error))
	setupIncludeWrapperFunc := func(name string) {
		funcMap[name] = func(data interface{}) (string, error) {
			return helmIncludeFunc(name, data)
		}
	}

	for _, name := range []string{} {
		setupIncludeWrapperFunc(name)
	}
}

func SetupWerfSecretFile(secretsRuntimeData *secrets.SecretsRuntimeData, funcMap template.FuncMap) {
	funcMap["werf_secret_file"] = func(secretRelativePath string) (string, error) {
		if path.IsAbs(secretRelativePath) {
			return "", fmt.Errorf("expected relative secret file path, given path %v", secretRelativePath)
		}

		decodedData, ok := secretsRuntimeData.DecryptedSecretFilesData[secretRelativePath]

		if !ok {
			var secretFiles []string
			for key := range secretsRuntimeData.DecryptedSecretFilesData {
				secretFiles = append(secretFiles, key)
			}

			return "", fmt.Errorf("secret file %q not found, you may use one of the following: %q", secretRelativePath, strings.Join(secretFiles, "', '"))
		}

		return decodedData, nil
	}
}
