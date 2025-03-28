package host_cleaning

import (
	"context"
	"errors"
	"fmt"

	"github.com/gofrs/flock"

	chart "github.com/werf/common-go/pkg/lock"
	"github.com/werf/lockgate/pkg/file_lock"
	"github.com/werf/lockgate/pkg/file_locker"
)

// AutoHostCleanup host-locking mechanism.
//
// The mechanism's goal is to ensure no parallel background "werf host cleanup" processes.
//
// How it works:
// 1) The foreground process, e.g. "werf build", does locker.isLockFree() to get actual lock's state.
//    If the lock is free, we start a "werf host cleanup" background process.
// 2) The started background process ("werf host cleanup") does locker.WithLockOrNothing().
//    The method tries the lock again. If lock is free, it executes the callback and releases it after.
//    If lock isn't free, the method does nothing.

type autoHostCleanupLocker struct {
	lock *flock.Flock

	lockName string
}

func newAutoHostCleanupLocker() (*autoHostCleanupLocker, error) {
	locksDir, err := werfLocksDir()
	if err != nil {
		return nil, fmt.Errorf("unable to get werf's locks dir: %w", err)
	}

	lockName := "werf.auto-host-cleanup"
	lockPath := werfLockPath(locksDir, lockName)

	return &autoHostCleanupLocker{
		lock:     flock.New(lockPath),
		lockName: lockName,
	}, nil
}

// isLockFree returns actual lock's state.
func (l *autoHostCleanupLocker) isLockFree(_ context.Context) (bool, error) {
	ok, err := l.lock.TryLock()
	if err != nil {
		return false, err
	}
	return ok, l.lock.Unlock()
}

// WithLockOrNothing tries auto-host-lock. If the lock is acquired, executes callback and releases the lock after.
// If the lock is NOT acquired, does nothing.
func (l *autoHostCleanupLocker) WithLockOrNothing(_ context.Context, callback func() error) error {
	locked, err := l.lock.TryLock()
	if err != nil {
		return err
	}
	if !locked {
		return nil
	}
	err = callback() // Should we handle panic here and release the lock anyway?
	if err != nil {
		return errors.Join(err, l.lock.Unlock()) // join non-nil errors or return nil
	}
	return nil
}

// werfLocksDir takes directory using common-go. Keeps compatibility with werf's locks.
func werfLocksDir() (string, error) {
	hostLocker, err := chart.HostLocker()
	if err != nil {
		return "", fmt.Errorf("unable to get host locker: %w", err)
	}

	return hostLocker.(*file_locker.FileLocker).LocksDir, nil
}

// werfLockPath takes lock's name using lockgate. Keeps compatibility with werf's locks.
func werfLockPath(locksDir, locksName string) string {
	lock := file_lock.NewFileLock(locksName, locksDir)
	return lock.(*file_lock.FileLock).LockFilePath()
}
