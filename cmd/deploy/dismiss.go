package main

import (
	"github.com/flant/dapp/pkg/deploy"
)

type dismissRubyCliOptions struct {
	Namespace     string `json:"namespace"`
	Context       string `json:"context"`
	WithNamespace bool   `json:"with_namespace"`
}

func runDismiss(releaseName, kubeContext string, rubyCliOptions dismissRubyCliOptions) error {
	return deploy.RunDismiss(releaseName, deploy.DismissOptions{
		Namespace:     rubyCliOptions.Namespace,
		KubeContext:   kubeContext,
		WithNamespace: rubyCliOptions.WithNamespace,
	})
}
