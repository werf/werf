package deploy

import (
	"github.com/flant/kubedog/pkg/kube"
	"github.com/flant/logboek"
	"github.com/flant/werf/pkg/deploy/helm"
	"github.com/flant/werf/pkg/storage"
)

type DismissOptions struct {
	WithNamespace bool
	WithHooks     bool
}

func RunDismiss(projectName, releaseName, namespace, _ string, storageLockManager storage.LockManager, opts DismissOptions) error {
	if lock, err := storageLockManager.LockDeployProcess(projectName, releaseName, kube.Context); err != nil {
		return err
	} else {
		defer storageLockManager.Unlock(lock)
	}

	logboek.Debug.LogF("Dismiss options: %#v\n", opts)
	logboek.Debug.LogF("Namespace: %s\n", namespace)
	return helm.PurgeHelmRelease(releaseName, namespace, opts.WithNamespace, opts.WithHooks)
}
