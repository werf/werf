package lock

import (
	"fmt"
	"time"
)

type Base struct {
	Name        string
	ActiveLocks int
}

type locker interface {
	Lock() error
	Unlock() error
}

type baseLocker struct {
	Timeout  time.Duration
	ReadOnly bool
	OnWait   func(doWait func() error) error
}

func (locker *baseLocker) Lock() error {
	return fmt.Errorf("baseLocker.Lock not implemented")
}

func (locker *baseLocker) Unlock() error {
	return fmt.Errorf("baseLocker.Unlock not implemented")
}

func (lock *Base) GetName() string {
	return lock.Name
}

func (lock *Base) Lock(l locker) error {
	if lock.ActiveLocks == 0 {
		err := l.Lock()
		if err != nil {
			return err
		}
	}

	lock.ActiveLocks += 1

	return nil
}

func (lock *Base) Unlock(l locker) error {
	if lock.ActiveLocks == 0 {
		return nil
	}

	lock.ActiveLocks -= 1

	if lock.ActiveLocks == 0 {
		return l.Unlock()
	}

	return nil
}

func (lock *Base) WithLock(locker locker, f func() error) error {
	var err error

	err = lock.Lock(locker)
	if err != nil {
		return err
	}

	resErr := f()

	err = lock.Unlock(locker)
	if err != nil {
		return err
	}

	return resErr
}
