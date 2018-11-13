package lock

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/flant/dapp/pkg/dapp"
)

var (
	Locks          map[string]LockObject
	DefaultTimeout = 24 * time.Hour
	LocksDir       = filepath.Join(dapp.GetHomeDir(), "locks")
)

func Init() error {
	Locks = make(map[string]LockObject)

	err := os.MkdirAll(LocksDir, 0755)
	if err != nil {
		return fmt.Errorf("cannot initialize locks dir: %s", err)
	}

	return nil
}

type LockOptions struct {
	Timeout  time.Duration
	ReadOnly bool
}

func Lock(name string, opts LockOptions) error {
	lock := getLock(name)

	return lock.Lock(
		getTimeout(opts), opts.ReadOnly,
		func(doWait func() error) error { return onWait(name, doWait) },
	)
}

func Unlock(name string) error {
	if _, hasKey := Locks[name]; !hasKey {
		return fmt.Errorf("no such lock `%s` found", name)
	}

	lock := getLock(name)

	return lock.Unlock()
}

func WithLock(name string, opts LockOptions, f func() error) error {
	lock := getLock(name)

	return lock.WithLock(
		getTimeout(opts), opts.ReadOnly,
		func(doWait func() error) error { return onWait(name, doWait) },
		f,
	)
}

func onWait(name string, doWait func() error) error {
	fmt.Printf("Waiting for locked resource `%s` ...\n", name)

	err := doWait()
	if err != nil {
		return err
	}

	fmt.Printf("Waiting for locked resource `%s` DONE\n", name)

	return err
}

func getTimeout(opts LockOptions) time.Duration {
	if opts.Timeout != 0 {
		return opts.Timeout
	}
	return DefaultTimeout
}

func getLock(name string) LockObject {
	if l, hasKey := Locks[name]; hasKey {
		return l
	}

	Locks[name] = NewFileLock(name, LocksDir)

	return Locks[name]
}

type LockObject interface {
	GetName() string
	Lock(timeout time.Duration, readOnly bool, onWait func(doWait func() error) error) error
	Unlock() error
	WithLock(timeout time.Duration, readOnly bool, onWait func(doWait func() error) error, f func() error) error
}
