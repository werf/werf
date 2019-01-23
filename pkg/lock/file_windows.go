// +build windows

package lock

import (
	"time"
)

func NewFileLock(name string, locksDir string) LockObject {
	return &File{Base: Base{Name: name}, LocksDir: locksDir}
}

type File struct {
	Base
	LocksDir string
}

func (lock *File) Lock(timeout time.Duration, readOnly bool, onWait func(doWait func() error) error) error {
	return nil
}

func (lock *File) Unlock() error {
	return nil
}

func (lock *File) WithLock(timeout time.Duration, readOnly bool, onWait func(doWait func() error) error, f func() error) error {
	return f()
}
