package werf

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/werf/lockgate"
)

func getWerfLastRunAtPath() string {
	return filepath.Join(GetServiceDir(), "var", "last_werf_run_at")
}

func getWerfFirstRunAtPath() string {
	return filepath.Join(GetServiceDir(), "var", "first_werf_run_at")
}

func SetWerfLastRunAt(ctx context.Context) error {
	path := getWerfLastRunAtPath()
	if _, lock, err := AcquireHostLock(ctx, path, lockgate.AcquireOptions{OnWaitFunc: func(lockName string, doWait func() error) error { return doWait() }}); err != nil {
		return fmt.Errorf("error locking path %q: %s", path, err)
	} else {
		defer ReleaseHostLock(lock)
	}

	return writeTimestampFile(path, time.Now())
}

func GetWerfLastRunAt(ctx context.Context) (time.Time, error) {
	path := getWerfLastRunAtPath()
	if _, lock, err := AcquireHostLock(ctx, path, lockgate.AcquireOptions{OnWaitFunc: func(lockName string, doWait func() error) error { return doWait() }}); err != nil {
		return time.Time{}, fmt.Errorf("error locking path %q: %s", path, err)
	} else {
		defer ReleaseHostLock(lock)
	}

	return readTimestampFile(path)
}

func SetWerfFirstRunAt(ctx context.Context) error {
	path := getWerfFirstRunAtPath()
	if _, lock, err := AcquireHostLock(ctx, path, lockgate.AcquireOptions{OnWaitFunc: func(lockName string, doWait func() error) error { return doWait() }}); err != nil {
		return fmt.Errorf("error locking path %q: %s", path, err)
	} else {
		defer ReleaseHostLock(lock)
	}

	if exists, err := checkTimestampFileExists(path); err != nil {
		return fmt.Errorf("error checking existance of %q: %s", path, err)
	} else if !exists {
		return writeTimestampFile(path, time.Now())
	}
	return nil
}

func GetWerfFirstRunAt(ctx context.Context) (time.Time, error) {
	path := getWerfFirstRunAtPath()
	if _, lock, err := AcquireHostLock(ctx, path, lockgate.AcquireOptions{OnWaitFunc: func(lockName string, doWait func() error) error { return doWait() }}); err != nil {
		return time.Time{}, fmt.Errorf("error locking path %q: %s", path, err)
	} else {
		defer ReleaseHostLock(lock)
	}

	return readTimestampFile(path)
}

func readTimestampFile(path string) (time.Time, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return time.Time{}, nil
	} else if err != nil {
		return time.Time{}, fmt.Errorf("error accessing %q: %s", path, err)
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return time.Time{}, fmt.Errorf("error reading %q: %s", path, err)
	}

	i, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		// os.RemoveAll(path)
		// return time.Time{}, nil
		return time.Time{}, fmt.Errorf("error parsing %q timestamp data %q: %s", path, data, err)
	}

	return time.Unix(i, 0), nil
}

func writeTimestampFile(path string, t time.Time) error {
	timeStr := fmt.Sprintf("%d\n", t.Unix())

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("error creating dir %q: %s", dir, err)
	}

	if err := ioutil.WriteFile(path, []byte(timeStr), 0644); err != nil {
		return fmt.Errorf("error writing %q: %s", path, err)
	}

	return nil
}

func checkTimestampFileExists(path string) (bool, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("error accessing %q: %s", path, err)
	}
	return true, nil
}
