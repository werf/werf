package deploy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flant/werf/pkg/deploy/secret"
)

func getSafeSecretManager(projectDir string, secretValues []string) (secret.Manager, error) {
	isSecretsExists := false
	if _, err := os.Stat(filepath.Join(projectDir, ProjectSecretDir)); !os.IsNotExist(err) {
		isSecretsExists = true
	}
	if _, err := os.Stat(filepath.Join(projectDir, ProjectDefaultSecretValuesFile)); !os.IsNotExist(err) {
		isSecretsExists = true
	}
	if len(secretValues) > 0 {
		isSecretsExists = true
	}
	if isSecretsExists {
		key, err := secret.GetSecretKey(projectDir)
		if err != nil {
			if strings.HasPrefix(err.Error(), "encryption key not found in") {
				fmt.Fprintln(os.Stderr, err)
			} else {
				return nil, err
			}
		} else {
			return secret.NewManager(key, secret.NewManagerOptions{})
		}
	}

	return secret.NewSafeManager()
}

func getWerfChart(projectDir string, m secret.Manager, values, secretValues, set, setString []string, serviceValues map[string]interface{}) (*WerfChart, error) {
	werfChart, err := GenerateWerfChart(projectDir, m)
	if err != nil {
		return nil, err
	}

	for _, path := range values {
		err = werfChart.SetValuesFile(path)
		if err != nil {
			return nil, err
		}
	}

	for _, path := range secretValues {
		err = werfChart.SetSecretValuesFile(path, m)
		if err != nil {
			return nil, err
		}
	}

	for _, set := range set {
		err = werfChart.SetValuesSet(set)
		if err != nil {
			return nil, err
		}
	}

	for _, setString := range setString {
		err = werfChart.SetValuesSetString(setString)
		if err != nil {
			return nil, err
		}
	}

	if serviceValues != nil {
		err = werfChart.SetValues(serviceValues)
		if err != nil {
			return nil, err
		}
	}

	if debug() {
		fmt.Printf("Werf chart: %#v\n", werfChart)
	}

	return werfChart, nil
}
