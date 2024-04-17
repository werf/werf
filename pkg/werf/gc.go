package werf

import "github.com/werf/lockgate/pkg/file_lock"

func GCHostLockerDir() error {
	return file_lock.GCLockFileDir(getHostLockerDir())
}
