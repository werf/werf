package werf

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

var (
	Version = "dev"
)

var (
	tmpDir, homeDir  string
	sharedContextDir string
	localCacheDir    string
	serviceDir       string
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

	return nil
}
