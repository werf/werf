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
// error.
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
				if subdir == "local" {
					orphanedGitDir, err := orphanedWorktreeOrigin(worktreeDir)
					if err != nil {
						errs = append(errs, err)
						continue
					}
					desc.orphanedGitDir = orphanedGitDir
				}
			}

			res = append(res, desc)
		}
	}

	return res, errors.Join(errs...)
}

func orphanedWorktreeOrigin(worktreeDir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(worktreeDir, "git_dir"))
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("read worktree origin in %q: %w", worktreeDir, err)
	}

	origin := strings.TrimSuffix(string(data), "\n")
	if !filepath.IsAbs(origin) {
		return "", nil
	}

	if _, err := os.Stat(origin); err == nil {
		return "", nil
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("inspect worktree origin %q: %w", origin, err)
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
