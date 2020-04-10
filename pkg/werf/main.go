package werf

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/flant/logboek"

	"github.com/flant/lockgate"
)

var (
	Version = "dev"
)

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

func GetStorageLockManagerDir() string {
	return filepath.Join(GetSharedContextDir(), "storage", "locks")
}

func GetHostLocker() lockgate.Locker {
	return hostLocker
}

func SetupHostLockerDefaultOptions(opts lockgate.AcquireOptions) lockgate.AcquireOptions {
	if opts.OnWaitFunc == nil {
		opts.OnWaitFunc = onHostLockerWaitFunc
	}
	return opts
}

func WithHostLock(lockName string, opts lockgate.AcquireOptions, f func() error) error {
	return lockgate.WithAcquire(GetHostLocker(), lockName, SetupHostLockerDefaultOptions(opts), func(_ bool) error {
		return f()
	})
}

func AcquireHostLock(lockName string, opts lockgate.AcquireOptions) (bool, error) {
	return GetHostLocker().Acquire(lockName, SetupHostLockerDefaultOptions(opts))
}

func ReleaseHostLock(lockName string) error {
	return GetHostLocker().Release(lockName)
}

func onHostLockerWaitFunc(lockName string, doWait func() error) error {
	logProcessMsg := fmt.Sprintf("Waiting for locked resource %q", lockName)
	return logboek.LogProcessInline(logProcessMsg, logboek.LogProcessInlineOptions{}, func() error {
		return doWait()
	})
}

func Init(tmpDirOption, homeDirOption string) error {
	if val, ok := os.LookupEnv("WERF_TMP_DIR"); ok {
		tmpDir = val
	} else if tmpDirOption != "" {
		tmpDir = tmpDirOption
	} else {
		tmpDir = os.TempDir()
	}

	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		dir, err := filepath.EvalSymlinks(tmpDir)
		if err != nil {
			return fmt.Errorf("eval symlinks of path %s failed: %s", tmpDir, err)
		}

		tmpDir = dir
	}

	if val, ok := os.LookupEnv("WERF_HOME"); ok {
		homeDir = val
	} else if homeDirOption != "" {
		homeDir = homeDirOption
	} else {
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

	if locker, err := lockgate.NewFileLocker(filepath.Join(serviceDir, "locks")); err != nil {
		return fmt.Errorf("error creating werf host file locker: %s", err)
	} else {
		hostLocker = locker
	}

	return nil
}
