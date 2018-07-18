package lock

import (
	"time"
)

type Lock interface {
	GetName() string
	Lock(timeout time.Duration, readOnly bool, onWait func(doWait func() error) error) error
	Unlock() error
	WithLock(timeout time.Duration, readOnly bool, onWait func(doWait func() error) error, f func() error) error
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
