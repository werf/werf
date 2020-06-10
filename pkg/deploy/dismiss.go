package deploy

import (
	"github.com/flant/logboek"
	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/werf/pkg/deploy/helm"
	"github.com/werf/werf/pkg/storage"
)

type DismissOptions struct {
	WithNamespace bool
	WithHooks     bool
}

func RunDismiss(projectName, release, namespace, _ string, storageLockManager storage.LockManager, opts DismissOptions) error {
	if lock, err := storageLockManager.LockDeployProcess(projectName, release, kube.Context); err != nil {
		return err
	} else {
		defer storageLockManager.Unlock(lock)
	}

	if err := logboek.Default.LogBlock("Deploy options", logboek.LevelLogBlockOptions{}, func() error {
		logboek.LogF("Kubernetes namespace: %s\n", namespace)
		logboek.LogF("Helm release storage namespace: %s\n", helm.HelmReleaseStorageNamespace)
		logboek.LogF("Helm release storage type: %s\n", helm.HelmReleaseStorageType)
		logboek.LogF("Helm release name: %s\n", release)

		return nil
	}); err != nil {
		return err
	}

	logboek.Debug.LogF("Dismiss options: %#v\n", opts)
	logboek.Debug.LogF("Namespace: %s\n", namespace)
	return helm.PurgeHelmRelease(release, namespace, opts.WithNamespace, opts.WithHooks)
}
