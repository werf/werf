package gitdata

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/werf/common-go/pkg/util/timestamps"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/volumeutils"
)

type GitWorktreeDesc struct {
	Path          string
	LastAccessAt  time.Time
	Size          uint64
	CacheBasePath string
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
func GetGitWorktreesAndRemoveInvalid(ctx context.Context, cacheVersionRoot string) ([]GitDataEntry, error) {
	var res []GitDataEntry

	for _, subdir := range []string{"local", "remote"} {
		dir := filepath.Join(cacheVersionRoot, subdir)

		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return nil, fmt.Errorf("error accessing dir %q: %w", dir, err)
		}

		worktreeDirs, err := ioutil.ReadDir(dir)
		if err != nil {
			return nil, fmt.Errorf("error reading dir %q: %w", dir, err)
		}

		for _, worktreeDirInfo := range worktreeDirs {
			worktreeDir := filepath.Join(dir, worktreeDirInfo.Name())

			if !worktreeDirInfo.IsDir() {
				logboek.Context(ctx).Warn().LogF("Removing invalid entry %q: not a directory\n", worktreeDir)
				if err := os.RemoveAll(worktreeDir); err != nil {
					return nil, fmt.Errorf("unable to remove %q: %w", worktreeDir, err)
				}
				continue
			}

			size, err := volumeutils.DirSizeBytes(worktreeDir)
			if err != nil {
				return nil, fmt.Errorf("error getting dir %q size: %w", worktreeDir, err)
			}

			lastAccessAtPath := filepath.Join(worktreeDir, "last_access_at")
			lastAccessAt, err := timestamps.ReadTimestampFile(lastAccessAtPath)
			if err != nil {
				logboek.Context(ctx).Warn().LogF("Removing invalid entry %q: unable to read last access timestamp file %q: %s\n", worktreeDir, lastAccessAtPath, err)
				if err := os.RemoveAll(worktreeDir); err != nil {
					return nil, fmt.Errorf("unable to remove %q: %w", worktreeDir, err)
				}

				continue
			}

			res = append(res, &GitWorktreeDesc{
				Path:          worktreeDir,
				Size:          size,
				LastAccessAt:  lastAccessAt,
				CacheBasePath: dir,
			})
		}
	}

	return res, nil
}
