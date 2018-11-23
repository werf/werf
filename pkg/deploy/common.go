package deploy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flant/dapp/pkg/deploy/secret"
	"github.com/flant/dapp/pkg/slug"
	"github.com/flant/kubedog/pkg/kube"
)

func getNamespace(namespaceOption string) string {
	if namespaceOption == "" {
		return kube.DefaultNamespace
	}
	return slug.Slug(namespaceOption)
}

func getOptionalSecret(projectDir string, secretValues []string) (secret.Secret, error) {
	var s secret.Secret

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
		var err error
		s, err = secret.GetSecret(projectDir)
		if err != nil {
			if strings.HasPrefix(err.Error(), "encryption key not found in") {
				fmt.Fprintln(os.Stderr, err)
			} else {
				return nil, err
			}
		}
	}

	return s, nil
}

func getDappChart(projectDir string, s secret.Secret, values, secretValues, set []string, serviceValues map[string]interface{}) (*DappChart, error) {
	dappChart, err := GenerateDappChart(projectDir, s)
	if err != nil {
		return nil, err
	}
	if debug() {
		// Do not remove tmp chart in debug
		fmt.Printf("Generated dapp chart: %#v\n", dappChart)
	} else {
		defer os.RemoveAll(dappChart.ChartDir)
	}

	for _, path := range values {
		err = dappChart.SetValuesFile(path)
		if err != nil {
			return nil, err
		}
	}

	for _, path := range secretValues {
		err = dappChart.SetSecretValuesFile(path, s)
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

	if serviceValues != nil {
		err = dappChart.SetValues(serviceValues)
		if err != nil {
			return nil, err
		}
	}

	return dappChart, nil
}
