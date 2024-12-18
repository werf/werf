package command_helpers_for_werf_helm

import (
	"context"

	lock_manager "github.com/werf/nelm-for-werf-helm/pkg/lock_manager"
)

func LockReleaseWrapper(
	ctx context.Context,
	releaseName string,
	lockManager *lock_manager.LockManager,
	cmdFunc func() error,
) error {
	if lock, err := lockManager.LockRelease(ctx, releaseName); err != nil {
		return err
	} else {
		defer lockManager.Unlock(lock)
	}

	return cmdFunc()
}
