package werf

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/hofstadter-io/cinful"

	"github.com/werf/3p-helm/pkg/chart/loader"
	"github.com/werf/common-go/pkg/locker"
	"github.com/werf/common-go/pkg/secrets_manager"
	"github.com/werf/lockgate/pkg/file_lock"
	secrets_manager_legacy "github.com/werf/nelm-for-werf-helm/pkg/secrets_manager"
)

var (
	Version = "dev"
	Domain  = "werf.io"
)

var (
	tmpDir  string
	homeDir string

	sharedContextDir string
	localCacheDir    string
	serviceDir       string

	hostLocker *locker.HostLocker
)

func HostLocker() *locker.HostLocker {
	if hostLocker == nil {
		panic("bug: init required!")
	}
	return hostLocker
}

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

// Init initialize variables, locks, secrets, etc.
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
			return fmt.Errorf("eval symlinks of path %s failed: %w", tmpDir, err)
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
			return fmt.Errorf("get user home dir failed: %w", err)
		}

		homeDir = filepath.Join(userHomeDir, ".werf")
	}

	sharedContextDir = filepath.Join(homeDir, "shared_context")
	localCacheDir = filepath.Join(homeDir, "local_cache")
	serviceDir = filepath.Join(homeDir, "service")

	// TODO: options + update purgeHomeWerfFiles

	loader.SetLocalCacheDir(GetLocalCacheDir())
	loader.SetServiceDir(GetServiceDir())
	secrets_manager.SetWerfHomeDir(GetHomeDir())
	secrets_manager_legacy.WerfHomeDir = GetHomeDir()

	file_lock.LegacyHashFunction = true

	var err error
	if hostLocker, err = locker.NewHostLocker(filepath.Join(GetServiceDir(), "locks")); err != nil {
		return fmt.Errorf("error creating werf host file locker: %w", err)
	}

	if err := SetWerfFirstRunAt(context.Background()); err != nil {
		return fmt.Errorf("error setting werf first run at timestamp: %w", err)
	}

	if err := SetWerfLastRunAt(context.Background()); err != nil {
		return fmt.Errorf("error setting werf last run at timestamp: %w", err)
	}

	switch v := os.Getenv("WERF_STAGED_DOCKERFILE_VERSION"); v {
	case "", "v1":
		stagedDockerfileVersion = StagedDockerfileV1
	case "v2":
		stagedDockerfileVersion = StagedDockerfileV2
	default:
		return fmt.Errorf("unsupported WERF_STAGED_DOCKERFILE_VERSION=%q, expected v1 or v2 (recommended)", v)
	}

	return nil
}

func IsRunningInCI() bool {
	if ciInfo := cinful.Info(); ciInfo != nil {
		return true
	}
	return false
}
