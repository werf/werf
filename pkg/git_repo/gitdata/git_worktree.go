package gitdata

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/werf/logboek"
	"github.com/werf/werf/v3/pkg/volumeutils"
)

type GitWorktreeDesc struct {
	Path           string
	LastAccessAt   time.Time
	Size           uint64
	CacheBasePath  string
	HasSubmodules  bool
	orphanedGitDir string
}

func (entry *GitWorktreeDesc) GetPaths() []string {
	return []string{entry.Path}
}

func (entry *GitWorktreeDesc) GetSize() uint64 {
	return entry.Size
}

func (entry *GitWorktreeDesc) GetLastAccessAt() time.Time {
	return entry.LastAccessAt
}

func (entry *GitWorktreeDesc) GetCacheBasePath() string {
	return entry.CacheBasePath
}

// GetGitWorktreesAndRemoveInvalid scans the given cacheVersionRoot directory and returns
// a list of GitWorktreeDesc for each valid git worktree found. It removes invalid
// entries and handles errors appropriately.
//
// The directory structure expected is as follows:
// ├── 9/
// │   ├── local/
// │   │   ├── <worktree_hash>/
// │   │   │   └── ... (repository files)
// │   │   └── ... (other worktrees)
// │   ├── remote/
// │   │   ├── <worktree_hash>/
// │   │   │   └── ... (repository files)
// │   │   └── ... (other worktrees)
// └── ... (other cache versions)
//
// An entry whose size, access marker or recorded origin cannot be read is
// preserved and left out of the result; its error is joined into the returned
// error. Whether a recorded origin still exists is NOT probed here: that answer
// comes from ProbeMissingLocalWorktreeOrigins, which runs before the GC lock.
func GetGitWorktreesAndRemoveInvalid(ctx context.Context, cacheVersionRoot string, options ScanOptions) ([]GitDataEntry, error) {
	var res []GitDataEntry
	var errs []error

	for _, subdir := range []string{"local", "remote"} {
		dir := filepath.Join(cacheVersionRoot, subdir)

		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return nil, fmt.Errorf("error accessing dir %q: %w", dir, err)
		}

		worktreeDirs, err := ioutil.ReadDir(dir)
		if err != nil {
			errs = append(errs, fmt.Errorf("read dir %q: %w", dir, err))
			continue
		}

		for _, worktreeDirInfo := range worktreeDirs {
			worktreeDir := filepath.Join(dir, worktreeDirInfo.Name())

			if !worktreeDirInfo.IsDir() {
				logboek.Context(ctx).Warn().LogF("Removing invalid entry %q: not a directory\n", worktreeDir)
				if err := removePath(worktreeDir, options); err != nil {
					return nil, fmt.Errorf("unable to remove %q: %w", worktreeDir, err)
				}
				continue
			}

			size, err := volumeutils.DirSizeBytes(worktreeDir)
			if err != nil {
				errs = append(errs, fmt.Errorf("get dir %q size: %w", worktreeDir, err))
				continue
			}

			lastAccessAtPath := filepath.Join(worktreeDir, "last_access_at")
			lastAccessAt, err := readLastAccessAt(lastAccessAtPath)
			if err != nil {
				errs = append(errs, fmt.Errorf("read worktree access timestamp %q: %w", lastAccessAtPath, err))
				continue
			}

			desc := &GitWorktreeDesc{
				Path:          worktreeDir,
				Size:          size,
				LastAccessAt:  lastAccessAt,
				CacheBasePath: dir,
			}

			if !shouldPreserveGitDataEntryByLru(desc) {
				desc.HasSubmodules = worktreeHasSubmodules(ctx, worktreeDir)
			}

			if subdir == "local" {
				origin, err := recordedWorktreeOrigin(worktreeDir)
				if err != nil {
					errs = append(errs, err)
					continue
				}
				if _, missing := options.MissingWorktreeOrigins[origin]; missing {
					desc.orphanedGitDir = origin
				}
			}

			res = append(res, desc)
		}
	}

	return res, errors.Join(errs...)
}

// ProbeMissingLocalWorktreeOrigins returns the recorded origins of LRU-eligible
// local worktrees that are gone from the filesystem. It is meant to run BEFORE
// the exclusive GC lock is taken: probing an origin on a hung network mount
// would otherwise block every git cache user for as long as the mount hangs.
// An origin that cannot be probed is left out — unknown is not orphaned, so the
// worktree is cleaned by ordinary LRU order instead.
func ProbeMissingLocalWorktreeOrigins(ctx context.Context, cacheVersionRoot string) map[string]struct{} {
	dir := filepath.Join(cacheVersionRoot, "local")

	worktreeDirs, err := ioutil.ReadDir(dir)
	if err != nil {
		if !os.IsNotExist(err) {
			logboek.Context(ctx).Warn().LogF("Unable to look up local worktree origins in %q: %s\n", dir, err)
		}
		return nil
	}

	res := map[string]struct{}{}

	for _, worktreeDirInfo := range worktreeDirs {
		if !worktreeDirInfo.IsDir() {
			continue
		}

		worktreeDir := filepath.Join(dir, worktreeDirInfo.Name())

		lastAccessAt, err := readLastAccessAt(filepath.Join(worktreeDir, "last_access_at"))
		if err != nil {
			continue
		}
		if shouldPreserveGitDataEntryByLru(&GitWorktreeDesc{LastAccessAt: lastAccessAt}) {
			continue
		}

		origin, err := recordedWorktreeOrigin(worktreeDir)
		if err != nil || origin == "" {
			continue
		}

		if _, err := os.Stat(origin); os.IsNotExist(err) {
			res[origin] = struct{}{}
		} else if err != nil {
			logboek.Context(ctx).Warn().LogF("Treating worktree origin %q as present: unable to inspect it: %s\n", origin, err)
		}
	}

	return res
}

func recordedWorktreeOrigin(worktreeDir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(worktreeDir, "git_dir"))
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("read worktree origin in %q: %w", worktreeDir, err)
	}

	origin := strings.TrimSuffix(string(data), "\n")
	if !filepath.IsAbs(origin) {
		return "", nil
	}

	return origin, nil
}

func worktreeHasSubmodules(ctx context.Context, worktreeCacheDir string) bool {
	_, err := os.Stat(filepath.Join(worktreeCacheDir, "worktree", ".gitmodules"))
	if err == nil {
		return true
	}
	if !os.IsNotExist(err) {
		logboek.Context(ctx).Warn().LogF("Treating %q as a worktree with submodules: unable to check for .gitmodules: %s\n", worktreeCacheDir, err)
		return true
	}
	return false
}
