package host_cleaning

import (
	"context"
	"errors"

	chart "github.com/werf/common-go/pkg/lock"
	"github.com/werf/lockgate"
	"github.com/werf/werf/v2/pkg/background"
)

type autoHostCleanupLocker struct {
	lockName    string
	lockOptions lockgate.AcquireOptions
	lockHandle  lockgate.LockHandle
}

func newAutoHostCleanupLocker() *autoHostCleanupLocker {
	return &autoHostCleanupLocker{
		lockName:    "werf.auto-host-cleanup",
		lockOptions: lockgate.AcquireOptions{NonBlocking: true},
	}
}

// TryLockOrNothing if "soft" (NonBlocking=true) auto-host-lock is acquired, executes callback.
// If callback returns false or err, releases the auto-host-lock.
// If "soft" lock is NOT acquired, does nothing.
func (l *autoHostCleanupLocker) TryLockOrNothing(ctx context.Context, callback func() (bool, error)) (bool, error) {
	var acquired bool
	var err error
	acquired, l.lockHandle, err = chart.AcquireHostLock(ctx, l.lockName, l.lockOptions)
	if err != nil {
		return false, err
	}
	if !acquired {
		return false, nil
	}
	ok, err := callback()
	if err != nil || !ok {
		return false, errors.Join(err, chart.ReleaseHostLock(l.lockHandle)) // join non-nil errors or return nil
	}
	return true, nil
}

// Unlock in background mode executes callback and releases auto-host-lock.
// In foreground mode just executes callback.
func (l *autoHostCleanupLocker) Unlock(ctx context.Context, callback func() error) error {
	if !background.IsBackgroundModeEnabled() {
		return callback()
	}

	var err error
	_, l.lockHandle, err = chart.AcquireHostLock(ctx, l.lockName, l.lockOptions) // build lock object
	if err != nil {
		return err
	}

	return errors.Join(callback(), chart.ReleaseHostLock(l.lockHandle)) // join non-nil errors or return nil
}
