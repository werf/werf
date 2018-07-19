package lock

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/flant/dapp/pkg/dapp"
	"github.com/spaolacci/murmur3"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

var (
	LocksPath = filepath.Join(dapp.HomeDir, "locks")
)

func NewFileLock(name string) Lock {
	err := os.MkdirAll(LocksPath, 0755)
	if err != nil {
		panic(fmt.Errorf("cannot create file lock: %s", err))
	}

	return &File{Base: Base{Name: name}}
}

type File struct {
	Base
	locker *fileLocker
}

func (lock *File) newLocker(timeout time.Duration, readOnly bool, onWait func(doWait func() error) error) *fileLocker {
	return &fileLocker{
		baseLocker: baseLocker{
			Timeout:  timeout,
			ReadOnly: readOnly,
			OnWait:   onWait,
		},
		FileLock: lock,
	}
}

func (lock *File) Lock(timeout time.Duration, readOnly bool, onWait func(doWait func() error) error) error {
	lock.locker = lock.newLocker(timeout, readOnly, onWait)
	return lock.Base.Lock(lock.locker)
}

func (lock *File) Unlock() error {
	if lock.locker == nil {
		return nil
	}

	err := lock.Base.Unlock(lock.locker)
	if err != nil {
		return err
	}

	lock.locker = nil

	return nil
}

func (lock *File) WithLock(timeout time.Duration, readOnly bool, onWait func(doWait func() error) error, f func() error) error {
	lock.locker = lock.newLocker(timeout, readOnly, onWait)

	err := lock.Base.WithLock(lock.locker, f)
	if err != nil {
		return err
	}

	lock.locker = nil

	return nil
}

type fileLocker struct {
	baseLocker

	FileLock        *File
	openFileHandler *os.File
}

func (locker *fileLocker) lockFilePath() string {
	h32 := murmur3.New32()
	h32.Write([]byte(locker.FileLock.GetName()))

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, h32.Sum32())
	if err != nil {
		panic(fmt.Errorf("cannot make lock file path for lock %s: %s", locker.FileLock.GetName(), err))
	}

	fileName := fmt.Sprintf("%x", buf.Bytes())

	return filepath.Join(LocksPath, fileName)
}

func (locker *fileLocker) Lock() error {
	f, err := os.OpenFile(locker.lockFilePath(), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	locker.openFileHandler = f

	fd := int(locker.openFileHandler.Fd())
	var mode int
	if locker.ReadOnly {
		mode = syscall.LOCK_SH
	} else {
		mode = syscall.LOCK_EX
	}

	err = syscall.Flock(fd, mode|syscall.LOCK_NB)

	if err == syscall.EWOULDBLOCK {
		return locker.OnWait(func() error {
			return locker.pollFlock(fd, mode)
		})
	}

	return err
}

func (locker *fileLocker) pollFlock(fd int, mode int) error {
	flockRes := make(chan error)
	cancelPoll := make(chan bool)

	go func() {
		ticker := time.NewTicker(time.Millisecond * 500)

	PollFlock:
		for {
			select {
			case <-ticker.C:
				err := syscall.Flock(fd, mode|syscall.LOCK_NB)
				if err == nil || err != syscall.EWOULDBLOCK {
					flockRes <- err
				}
			case <-cancelPoll:
				break PollFlock
			}
		}
	}()

	select {
	case err := <-flockRes:
		return err
	case <-time.After(locker.Timeout):
		cancelPoll <- true
		return fmt.Errorf("lock `%s` timeout %s expired", locker.FileLock.GetName(), locker.Timeout)
	}
}

func (locker *fileLocker) Unlock() error {
	err := locker.openFileHandler.Close()
	if err != nil {
		return err
	}

	locker.openFileHandler = nil

	return nil
}
