package main

import (
	"github.com/flant/dapp/pkg/deploy"
)

type lintRubyCliOptions struct {
	HelmSetOptions          []string `json:"helm_set_options"`
	HelmValuesOptions       []string `json:"helm_values_options"`
	HelmSecretValuesOptions []string `json:"helm_secret_values_options"`
}

func runLint(projectDir string, dimgs []*deploy.DimgInfoGetterStub, rubyCliOptions lintRubyCliOptions) error {
	return deploy.RunLint(deploy.LintOptions{
		ProjectDir:   projectDir,
		Values:       rubyCliOptions.HelmValuesOptions,
		SecretValues: rubyCliOptions.HelmSecretValuesOptions,
		Set:          rubyCliOptions.HelmSetOptions,
		Dimgs:        dimgs,
	})
}
