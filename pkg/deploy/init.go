package deploy

import "github.com/flant/werf/pkg/deploy/helm"

func Init(kubeContext string) error {
	helm.KubeContext = kubeContext
	return helm.ValidateHelmVersion()
}
