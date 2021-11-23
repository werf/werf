package werf

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/werf/lockgate/pkg/file_lock"
	"github.com/werf/lockgate/pkg/file_locker"
	"github.com/werf/logboek"

	"github.com/werf/lockgate"
)

var Version = "dev"

var (
	tmpDir, homeDir  string
	sharedContextDir string
	localCacheDir    string
	serviceDir       string

	hostLocker lockgate.Locker
)

func GetSharedContextDir() string {
	if sharedContextDir == "" {
		panic("bug: init required!")
	}

	return sharedContextDir
}

func GetLocalCacheDir() string {
	if localCacheDir == "" {
		panic("bug: init required!")
	}

	return localCacheDir
}

func GetServiceDir() string {
	if serviceDir == "" {
		panic("bug: init required!")
	}

	return serviceDir
}

func GetHomeDir() string {
	if homeDir == "" {
		panic("bug: init required!")
	}

	return homeDir
}

func GetTmpDir() string {
	if tmpDir == "" {
		panic("bug: init required!")
	}

	return tmpDir
}

func GetStagesStorageCacheDir() string {
	return filepath.Join(GetSharedContextDir(), "storage", "stages_storage_cache", "1")
}

func GetHostLocker() lockgate.Locker {
	return hostLocker
}

func SetupLockerDefaultOptions(ctx context.Context, opts lockgate.AcquireOptions) lockgate.AcquireOptions {
	if opts.OnWaitFunc == nil {
		opts.OnWaitFunc = DefaultLockerOnWait(ctx)
	}
	if opts.OnLostLeaseFunc == nil {
		opts.OnLostLeaseFunc = DefaultLockerOnLostLease
	}
	return opts
}

func WithHostLock(ctx context.Context, lockName string, opts lockgate.AcquireOptions, f func() error) error {
	return lockgate.WithAcquire(GetHostLocker(), lockName, SetupLockerDefaultOptions(ctx, opts), func(_ bool) error {
		return f()
	})
}

func AcquireHostLock(ctx context.Context, lockName string, opts lockgate.AcquireOptions) (bool, lockgate.LockHandle, error) {
	return GetHostLocker().Acquire(lockName, SetupLockerDefaultOptions(ctx, opts))
}

func ReleaseHostLock(lock lockgate.LockHandle) error {
	return GetHostLocker().Release(lock)
}

func DefaultLockerOnWait(ctx context.Context) func(lockName string, doWait func() error) error {
	return func(lockName string, doWait func() error) error {
		logProcessMsg := fmt.Sprintf("Waiting for locked %q", lockName)
		return logboek.Context(ctx).LogProcessInline(logProcessMsg).DoError(doWait)
	}
}

func DefaultLockerOnLostLease(lock lockgate.LockHandle) error {
	panic(fmt.Sprintf("Locker has lost lease for locked %q uuid %s. Will crash current process immediately!", lock.LockName, lock.UUID))
}

func Init(tmpDirOption, homeDirOption string) error {
	val, ok := os.LookupEnv("WERF_TMP_DIR")
	switch {
	case ok:
		tmpDir = val
	case tmpDirOption != "":
		tmpDir = tmpDirOption
	default:
		tmpDir = os.TempDir()
	}

	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		dir, err := filepath.EvalSymlinks(tmpDir)
		if err != nil {
			return fmt.Errorf("eval symlinks of path %s failed: %s", tmpDir, err)
		}

		tmpDir = dir
	}

	val, ok = os.LookupEnv("WERF_HOME")
	switch {
	case ok:
		homeDir = val
	case homeDirOption != "":
		homeDir = homeDirOption
	default:
		userHomeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("get user home dir failed: %s", err)
		}

		homeDir = filepath.Join(userHomeDir, ".werf")
	}

	// TODO: options + update purgeHomeWerfFiles

	sharedContextDir = filepath.Join(homeDir, "shared_context")
	localCacheDir = filepath.Join(homeDir, "local_cache")
	serviceDir = filepath.Join(homeDir, "service")

	file_lock.LegacyHashFunction = true

	if locker, err := file_locker.NewFileLocker(filepath.Join(serviceDir, "locks")); err != nil {
		return fmt.Errorf("error creating werf host file locker: %s", err)
	} else {
		hostLocker = locker
	}

	if err := SetWerfFirstRunAt(context.Background()); err != nil {
		return fmt.Errorf("error setting werf first run at timestamp: %s", err)
	}

	if err := SetWerfLastRunAt(context.Background()); err != nil {
		return fmt.Errorf("error setting werf last run at timestamp: %s", err)
	}

	return nil
}
