package deploy

import "github.com/flant/werf/pkg/deploy/helm"

func Init(kubeConfig, kubeContext, tillerNamespace, tillerStorage string) error {
	if err := helm.Init(kubeConfig, kubeContext, tillerNamespace, tillerStorage); err != nil {
		return err
	}

	return nil
}
