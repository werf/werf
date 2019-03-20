package deploy

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/flant/logboek"
	"github.com/flant/werf/pkg/deploy/secret"
	"github.com/flant/werf/pkg/deploy/werf_chart"
)

func GetSafeSecretManager(projectDir string, secretValues []string) (secret.Manager, error) {
	isSecretsExists := false
	if _, err := os.Stat(filepath.Join(projectDir, werf_chart.ProjectSecretDir)); !os.IsNotExist(err) {
		isSecretsExists = true
	}
	if _, err := os.Stat(filepath.Join(projectDir, werf_chart.ProjectDefaultSecretValuesFile)); !os.IsNotExist(err) {
		isSecretsExists = true
	}
	if len(secretValues) > 0 {
		isSecretsExists = true
	}
	if isSecretsExists {
		key, err := secret.GetSecretKey(projectDir)
		if err != nil {
			if strings.HasPrefix(err.Error(), "encryption key not found in") {
				logboek.LogErrorF("WARNING: Unable to get secrets key: %s\n", err)
			} else {
				return nil, err
			}
		} else {
			return secret.NewManager(key, secret.NewManagerOptions{})
		}
	}

	return secret.NewSafeManager()
}
