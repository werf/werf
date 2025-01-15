package lock_manager

import (
	"context"
	"fmt"

	"github.com/werf/common-go/pkg/lock"
	"github.com/werf/lockgate"
	"github.com/werf/logboek"
)

func NewGeneric(locker lockgate.Locker) *Generic {
	return &Generic{Locker: locker}
}

type Generic struct {
	// Single Locker for all projects
	Locker lockgate.Locker
}

func (manager *Generic) LockStage(ctx context.Context, projectName, digest string) (LockHandle, error) {
	_, lock, err := manager.Locker.Acquire(genericStageLockName(projectName, digest), chart.SetupLockerDefaultOptions(ctx, lockgate.AcquireOptions{}))
	return LockHandle{LockgateHandle: lock, ProjectName: projectName}, err
}

func (manager *Generic) Unlock(ctx context.Context, lock LockHandle) error {
	err := manager.Locker.Release(lock.LockgateHandle)
	if err != nil {
		logboek.Context(ctx).Error().LogF("ERROR: unable to release lock for %q: %s\n", lock.LockgateHandle.LockName, err)
	}
	return err
}

func genericStageLockName(projectName, digest string) string {
	return fmt.Sprintf("%s.%s", projectName, digest)
}
