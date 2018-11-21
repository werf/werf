package deploy

import (
	"time"

	"github.com/flant/dapp/pkg/secret"
)

type DeployOptions struct {
	Values       []string
	SecretValues []string
	Set          []string
	Secret       secret.Secret
	Timeout      time.Duration
	KubeContext  string
}

func RunDeploy(projectDir string, releaseName string, namespace string, opts DeployOptions) error {
	dappChart, err := GenerateDappChart(projectDir, opts.Secret)
	if err != nil {
		return err
	}

	for _, path := range opts.Values {
		err = dappChart.SetValuesFile(path)
		if err != nil {
			return err
		}
	}

	for _, path := range opts.SecretValues {
		err = dappChart.SetSecretValuesFile(path, opts.Secret)
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
