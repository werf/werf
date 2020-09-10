package deploy

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/deploy/lock_manager"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/deploy/helm"
	"github.com/werf/werf/pkg/kubeutils"
)

type DismissOptions struct {
	WithNamespace bool
	WithHooks     bool
}

func RunDismiss(ctx context.Context, projectName, release, namespace, _ string, opts DismissOptions) error {
	lockManager, err := lock_manager.NewLockManager(namespace)
	if err != nil {
		return err
	}

	if err := func() error {
		if lock, err := lockManager.LockRelease(ctx, release); err != nil {
			return err
		} else {
			defer lockManager.Unlock(lock)
		}

		if err := logboek.Context(ctx).Default().LogBlock("Deploy options").DoError(func() error {
			logboek.Context(ctx).LogF("Kubernetes namespace: %s\n", namespace)
			logboek.Context(ctx).LogF("Helm release storage namespace: %s\n", helm.HelmReleaseStorageNamespace)
			logboek.Context(ctx).LogF("Helm release storage type: %s\n", helm.HelmReleaseStorageType)
			logboek.Context(ctx).LogF("Helm release name: %s\n", release)

			return nil
		}); err != nil {
			return err
		}

		logboek.Context(ctx).Debug().LogF("Dismiss options: %#v\n", opts)
		logboek.Context(ctx).Debug().LogF("Namespace: %s\n", namespace)
		return helm.PurgeHelmRelease(ctx, release, namespace, opts.WithHooks)
	}(); err != nil {
		return err
	}

	if opts.WithNamespace {
		if err := kubeutils.RemoveResourceAndWaitUntilRemoved(ctx, namespace, "Namespace", ""); err != nil {
			return fmt.Errorf("delete namespace %s failed: %s", namespace, err)
		}
		logboek.Context(ctx).LogOptionalLn()
	}

	return nil
}
