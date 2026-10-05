package lock_manager

import "github.com/werf/lockgate"

type observedTestLocker struct {
	onAcquire func()
	err       error
}

var _ lockgate.Locker = (*observedTestLocker)(nil)

func (l *observedTestLocker) Acquire(name string, _ lockgate.AcquireOptions) (bool, lockgate.LockHandle, error) {
	l.onAcquire()
	return l.err == nil, lockgate.LockHandle{LockName: name}, l.err
}

func (l *observedTestLocker) Release(_ lockgate.LockHandle) error {
	return nil
}
