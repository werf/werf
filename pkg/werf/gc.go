package werf

import (
	"os"
	"path/filepath"
	"time"

	"github.com/werf/lockgate/pkg/file_lock"
)

// HostLocksGCMinAge is the minimum age a host lock file must reach before GC may
// remove it. Host locks live for seconds during normal runs, so a conservative
// window guarantees GC never touches an active lock while still reclaiming the
// long-lived stale files that exhaust inodes on long-running runners.
const HostLocksGCMinAge = 24 * time.Hour

// hostLocksAutoGCThreshold is the number of stale host lock files that triggers
// automatic host cleanup. Lock files are empty, so they never move the byte-based
// volume-usage thresholds; without this signal GC would never fire on the exact
// inode-exhaustion failure mode this addresses.
const hostLocksAutoGCThreshold = 10000

// hostLockerDirSchemaVersion namespaces the host locker directory so that
// older werf binaries (which acquire locks without the inode-safe retry
// against GCLockFileDir) never share a locks directory with a GC-capable
// werf. Mixing the two on the same host could otherwise let a new process's
// GC pass unlink a lock file an old process just opened, right before the
// old process takes its flock — the old acquire path has no way to detect
// the resulting dead inode. Bump this whenever the locking protocol changes
// in an incompatible way.
const hostLockerDirSchemaVersion = "2"

func getHostLockerDir() string {
	return filepath.Join(GetServiceDir(), "locks", hostLockerDirSchemaVersion)
}

// GCHostLockerDir removes stale host lock files older than HostLocksGCMinAge.
func GCHostLockerDir() error {
	return file_lock.GCLockFileDir(getHostLockerDir(), HostLocksGCMinAge)
}

// ShouldRunHostLocksGC reports whether the host locks dir holds enough GC-eligible
// (older than HostLocksGCMinAge) files to warrant a cleanup. It scans lazily and
// stops as soon as the threshold is reached, so it stays cheap even on dirs with
// millions of entries.
func ShouldRunHostLocksGC() (bool, error) {
	dir := getHostLockerDir()
	f, err := os.Open(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	defer f.Close()

	cutoff := time.Now().Add(-HostLocksGCMinAge)
	count := 0
	for {
		entries, err := f.ReadDir(1024)
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			info, ierr := entry.Info()
			if ierr != nil {
				continue
			}
			if info.ModTime().Before(cutoff) {
				count++
				if count >= hostLocksAutoGCThreshold {
					return true, nil
				}
			}
		}
		if err != nil {
			break // io.EOF or read error: stop scanning
		}
	}

	return false, nil
}
