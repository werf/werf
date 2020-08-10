package deploy

import (
	"context"
	"fmt"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/deploy/helm"
	"github.com/werf/werf/pkg/kubeutils"
)

type DismissOptions struct {
	WithNamespace bool
	WithHooks     bool
}

func RunDismiss(ctx context.Context, projectName, release, namespace, _ string, opts DismissOptions) error {
	lockManager, err := NewLockManager(namespace)
	if err != nil {
		return err
	}

	if err := func() error {
		if lock, err := lockManager.LockRelease(release); err != nil {
			return err
		} else {
			defer lockManager.Unlock(lock)
		}

		if err := logboek.Default().LogBlock("Deploy options").DoError(func() error {
			logboek.LogF("Kubernetes namespace: %s\n", namespace)
			logboek.LogF("Helm release storage namespace: %s\n", helm.HelmReleaseStorageNamespace)
			logboek.LogF("Helm release storage type: %s\n", helm.HelmReleaseStorageType)
			logboek.LogF("Helm release name: %s\n", release)

			return nil
		}); err != nil {
			return err
		}

		logboek.Debug().LogF("Dismiss options: %#v\n", opts)
		logboek.Debug().LogF("Namespace: %s\n", namespace)
		return helm.PurgeHelmRelease(ctx, release, namespace, opts.WithHooks)
	}(); err != nil {
		return err
	}

	if opts.WithNamespace {
		if err := kubeutils.RemoveResourceAndWaitUntilRemoved(namespace, "Namespace", ""); err != nil {
			return fmt.Errorf("delete namespace %s failed: %s", namespace, err)
		}
		logboek.LogOptionalLn()
	}

	return nil
}
