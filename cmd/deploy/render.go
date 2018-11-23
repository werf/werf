package main

import (
	"github.com/flant/dapp/pkg/deploy"
)

type renderRubyCliOptions struct {
	HelmSetOptions          []string `json:"helm_set_options"`
	HelmValuesOptions       []string `json:"helm_values_options"`
	HelmSecretValuesOptions []string `json:"helm_secret_values_options"`
}

func runRender(projectDir string, rubyCliOptions renderRubyCliOptions) error {
	return deploy.RunRender(deploy.RenderOptions{
		ProjectDir:   projectDir,
		Values:       rubyCliOptions.HelmValuesOptions,
		SecretValues: rubyCliOptions.HelmSecretValuesOptions,
		Set:          rubyCliOptions.HelmSetOptions,
	})
}
