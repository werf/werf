package lock

import (
	"path/filepath"
	"time"

	"github.com/flant/werf/pkg/util"
)

func NewFileLock(name string, locksDir string) LockObject {
	return &FileLock{BaseLock: BaseLock{Name: name}, LocksDir: locksDir}
}

type FileLock struct {
	BaseLock
	LocksDir string
	locker   *fileLocker
}

func (lock *FileLock) newLocker(timeout time.Duration, readOnly bool, onWait func(doWait func() error) error) *fileLocker {
	return &fileLocker{
		baseLocker: baseLocker{
			Timeout:  timeout,
			ReadOnly: readOnly,
			OnWait:   onWait,
		},
		FileLock: lock,
	}
}

func (lock *FileLock) Lock(timeout time.Duration, readOnly bool, onWait func(doWait func() error) error) error {
	lock.locker = lock.newLocker(timeout, readOnly, onWait)
	return lock.BaseLock.Lock(lock.locker)
}

func (lock *FileLock) Unlock() error {
	if lock.locker == nil {
		return nil
	}

	err := lock.BaseLock.Unlock(lock.locker)
	if err != nil {
		return err
	}

	lock.locker = nil

	return nil
}

func (lock *FileLock) WithLock(timeout time.Duration, readOnly bool, onWait func(doWait func() error) error, f func() error) error {
	lock.locker = lock.newLocker(timeout, readOnly, onWait)

	err := lock.BaseLock.WithLock(lock.locker, f)
	if err != nil {
		return err
	}

	lock.locker = nil

	return nil
}

func (lock *FileLock) LockFilePath() string {
	fileName := util.MurmurHash(lock.GetName())
	return filepath.Join(lock.LocksDir, fileName)
}
