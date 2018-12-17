package deploy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flant/dapp/pkg/deploy/secret"
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

func getDappChart(projectDir string, m secret.Manager, values, secretValues, set, setString []string, serviceValues map[string]interface{}) (*DappChart, error) {
	dappChart, err := GenerateDappChart(projectDir, m)
	if err != nil {
		return nil, err
	}

	for _, path := range values {
		err = dappChart.SetValuesFile(path)
		if err != nil {
			return nil, err
		}
	}

	for _, path := range secretValues {
		err = dappChart.SetSecretValuesFile(path, m)
		if err != nil {
			return nil, err
		}
	}

	for _, set := range set {
		err = dappChart.SetValuesSet(set)
		if err != nil {
			return nil, err
		}
	}

	for _, setString := range setString {
		err = dappChart.SetValuesSetString(setString)
		if err != nil {
			return nil, err
		}
	}

	if serviceValues != nil {
		err = dappChart.SetValues(serviceValues)
		if err != nil {
			return nil, err
		}
	}

	if debug() {
		fmt.Printf("Dapp chart: %#v\n", dappChart)
	}

	return dappChart, nil
}
