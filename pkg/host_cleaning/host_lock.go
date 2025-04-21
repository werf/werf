package host_cleaning

import (
	"context"
	"errors"

	"github.com/werf/lockgate"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/werf"
)

// withHostLockOrNothing tries a lock. If the lock is acquired, executes callback and releases the lock after.
// If the lock is NOT acquired, does nothing.
func withHostLockOrNothing(ctx context.Context, lockName string, callback func() error) error {
	lockOptions := lockgate.AcquireOptions{NonBlocking: true}

	acquired, lock, err := werf.HostLocker().AcquireLock(ctx, lockName, lockOptions)
	if err != nil {
		return err
	}

	if !acquired {
		logboek.Context(ctx).Info().LogF("Ignore locked %s\n", lockName)
		return nil
	}

	// Should we handle panic here and release the lock anyway?
	return errors.Join(callback(), werf.HostLocker().ReleaseLock(lock)) // join non-nil errors or return nil
}
