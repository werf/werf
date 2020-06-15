package locker_with_retry

import (
	"math/rand"
	"time"

	"github.com/flant/lockgate"
	"github.com/flant/logboek"
)

type LockerWithRetry struct {
	Locker  lockgate.Locker
	Options LockerWithRetryOptions
}

type LockerWithRetryOptions struct {
	MaxAcquireAttempts int
	MaxReleaseAttempts int
}

func NewLockerWithRetry(locker lockgate.Locker, opts LockerWithRetryOptions) *LockerWithRetry {
	return &LockerWithRetry{Locker: locker, Options: opts}
}

func (locker *LockerWithRetry) Acquire(lockName string, opts lockgate.AcquireOptions) (acquired bool, handle lockgate.LockHandle, err error) {
	executeWithRetry(locker.Options.MaxAcquireAttempts, func() error {
		acquired, handle, err = locker.Locker.Acquire(lockName, opts)
		if err != nil {
			logboek.Error.LogF("ERROR: unable to acquire lock %s: %s\n", lockName, err)
		}
		return err
	})

	return
}

func (locker *LockerWithRetry) Release(lock lockgate.LockHandle) (err error) {
	executeWithRetry(locker.Options.MaxAcquireAttempts, func() error {
		err = locker.Locker.Release(lock)
		if err != nil {
			logboek.Error.LogF("ERROR: unable to release lock %s %s: %s\n", lock.UUID, lock.LockName, err)
		}
		return err
	})

	return
}

func executeWithRetry(maxAttempts int, executeFunc func() error) {
	attempt := 1

executeAttempt:
	if err := executeFunc(); err != nil {
		if attempt == maxAttempts {
			return
		}

		seconds := rand.Intn(10) // from 0 to 10 seconds
		logboek.Warn.LogF("Retrying in %d seconds (%d/%d) ...\n", seconds, attempt, maxAttempts)
		time.Sleep(time.Duration(seconds) * time.Second)

		attempt += 1
		goto executeAttempt
	}
}
