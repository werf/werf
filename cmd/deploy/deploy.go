package main

import (
	"time"

	"github.com/flant/dapp/pkg/deploy"
)

type deployRubyCliOptions struct {
	Namespace               string   `json:"namespace"`
	Repo                    string   `json:"repo"`
	Context                 string   `json:"context"`
	HelmSetOptions          []string `json:"helm_set_options"`
	HelmValuesOptions       []string `json:"helm_values_options"`
	HelmSecretValuesOptions []string `json:"helm_secret_values_options"`
	Timeout                 int64    `json:"timeout"`
	RegistryUsername        string   `json:"registry_username"`
	RegistryPassword        string   `json:"registry_password"`
	WithoutRegistry         bool     `json:"without_registry"`
}

func runDeploy(projectDir string, releaseName string, tag string, kubeContext string, repo string, rubyCliOptions deployRubyCliOptions) error {
	return deploy.RunDeploy(projectDir, releaseName, deploy.DeployOptions{
		Namespace:    rubyCliOptions.Namespace,
		Repo:         rubyCliOptions.Repo,
		Values:       rubyCliOptions.HelmValuesOptions,
		SecretValues: rubyCliOptions.HelmSecretValuesOptions,
		Set:          rubyCliOptions.HelmSetOptions,
		Timeout:      time.Duration(rubyCliOptions.Timeout) * time.Second,
		KubeContext:  kubeContext,
	})
}
