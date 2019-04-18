package deploy

import "github.com/flant/werf/pkg/deploy/helm"

func Init(kubeConfig, kubeContext, helmReleaseStorageNamespace, helmReleaseStorageType string) error {
	if err := helm.Init(kubeConfig, kubeContext, helmReleaseStorageNamespace, helmReleaseStorageType); err != nil {
		return err
	}

	return nil
}
