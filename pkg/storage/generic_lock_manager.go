package storage

import (
	"context"
	"fmt"

	"github.com/werf/lockgate"
	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/werf"
)

func NewGenericLockManager(locker lockgate.Locker) *GenericLockManager {
	return &GenericLockManager{Locker: locker}
}

type GenericLockManager struct {
	// Single Locker for all projects
	Locker lockgate.Locker
}

func (manager *GenericLockManager) LockStage(ctx context.Context, projectName, digest string) (LockHandle, error) {
	_, lock, err := manager.Locker.Acquire(genericStageLockName(projectName, digest), werf.SetupLockerDefaultOptions(ctx, lockgate.AcquireOptions{}))
	return LockHandle{LockgateHandle: lock, ProjectName: projectName}, err
}

func (manager *GenericLockManager) LockStageCache(ctx context.Context, projectName, digest string) (LockHandle, error) {
	_, lock, err := manager.Locker.Acquire(genericStageCacheLockName(projectName, digest), werf.SetupLockerDefaultOptions(ctx, lockgate.AcquireOptions{}))
	return LockHandle{LockgateHandle: lock, ProjectName: projectName}, err
}

func (manager *GenericLockManager) Unlock(ctx context.Context, lock LockHandle) error {
	err := manager.Locker.Release(lock.LockgateHandle)
	if err != nil {
		logboek.Context(ctx).Error().LogF("ERROR: unable to release lock for %q: %s\n", lock.LockgateHandle.LockName, err)
	}
	return err
}

func genericStageLockName(projectName, digest string) string {
	return fmt.Sprintf("%s.%s", projectName, digest)
}

func genericStageCacheLockName(projectName, digest string) string {
	return fmt.Sprintf("%s.%s.cache", projectName, digest)
}
