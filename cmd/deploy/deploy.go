package main

import "github.com/flant/dapp/pkg/deploy"

type deployRubyCliOptions struct {
	Namespace               string   `json:"namespace"`
	Context                 string   `json:"context"`
	HelmSetOptions          []string `json:"helm_set_options"`
	HelmValuesOptions       []string `json:"helm_values_options"`
	HelmSecretValuesOptions []string `json:"helm_secret_values_options"`
	Timeout                 int64    `json:"timeout"`
	RegistryUsername        string   `json:"registry_username"`
	RegistryPassword        string   `json:"registry_password"`
	WithoutRegistry         bool     `json:"without_registry"`
}

func runDeploy(projectDir string, releaseName string, tag string, kubeContext string, rubyCliOptions deployRubyCliOptions) error {
	return deploy.RunDeploy(projectDir, releaseName, rubyCliOptions.Namespace, deploy.DeployOptions{})
}
