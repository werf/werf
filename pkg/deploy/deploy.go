package deploy

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/flant/dapp/pkg/secret"
	"github.com/flant/dapp/pkg/slug"
)

type DeployOptions struct {
	Namespace    string
	Repo         string
	Values       []string
	SecretValues []string
	Set          []string
	Timeout      time.Duration
	KubeContext  string
}

// RunDeploy runs deploy of dapp chart
func RunDeploy(projectDir string, releaseName string, opts DeployOptions) error {
	namespace := slug.Slug(opts.Namespace)

	if debug() {
		fmt.Printf("Deploy options: %#v\n", opts)
		fmt.Printf("Slug namespace: %s\n", namespace)
	}

	var s secret.Secret

	isSecretsExists := false
	if _, err := os.Stat(filepath.Join(projectDir, ProjectSecretDir)); !os.IsNotExist(err) {
		isSecretsExists = true
	}
	if _, err := os.Stat(filepath.Join(projectDir, ProjectDefaultSecretValuesFile)); !os.IsNotExist(err) {
		isSecretsExists = true
	}
	if len(opts.SecretValues) > 0 {
		isSecretsExists = true
	}
	if isSecretsExists {
		var err error
		s, err = GetSecret(projectDir)
		if err != nil {
			return fmt.Errorf("cannot get project secret: %s", err)
		}
	}

	dappChart, err := GenerateDappChart(projectDir, s)
	if err != nil {
		return err
	}
	if debug() {
		// Do not remove tmp chart in debug
		fmt.Printf("Generated dapp chart: %#v\n", dappChart)
	} else {
		defer os.RemoveAll(dappChart.ChartDir)
	}

	for _, path := range opts.Values {
		err = dappChart.SetValuesFile(path)
		if err != nil {
			return err
		}
	}

	for _, path := range opts.SecretValues {
		err = dappChart.SetSecretValuesFile(path, s)
		if err != nil {
			return err
		}
	}

	for _, set := range opts.Set {
		err = dappChart.SetValuesSet(set)
		if err != nil {
			return err
		}
	}

	// TODO set service values

	return dappChart.Deploy(releaseName, namespace, HelmChartOptions{KubeContext: opts.KubeContext, Timeout: opts.Timeout})
}

func debug() bool {
	return os.Getenv("DAPP_DEPLOY_DEBUG") == "1"
}
