package gitdata

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/werf/werf/pkg/util/timestamps"
	"github.com/werf/werf/pkg/volumeutils"
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

func GetExistingGitWorktrees(cacheVersionRoot string) ([]*GitWorktreeDesc, error) {
	var res []*GitWorktreeDesc

	for _, dir := range []string{filepath.Join(cacheVersionRoot, "local"), filepath.Join(cacheVersionRoot, "remote")} {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return nil, fmt.Errorf("error accessing dir %q: %w", dir, err)
		}

		files, err := ioutil.ReadDir(dir)
		if err != nil {
			return nil, fmt.Errorf("error reading dir %q: %w", dir, err)
		}

		for _, finfo := range files {
			worktreePath := filepath.Join(dir, finfo.Name())

			if !finfo.IsDir() {
				if err := os.RemoveAll(worktreePath); err != nil {
					return nil, fmt.Errorf("unable to remove %q: %w", worktreePath, err)
				}
			}

			size, err := volumeutils.DirSizeBytes(worktreePath)
			if err != nil {
				return nil, fmt.Errorf("error getting dir %q size: %w", worktreePath, err)
			}

			lastAccessAtPath := filepath.Join(worktreePath, "last_access_at")
			lastAccessAt, err := timestamps.ReadTimestampFile(lastAccessAtPath)
			if err != nil {
				return nil, fmt.Errorf("error reading last access timestamp file %q: %w", lastAccessAtPath, err)
			}

			res = append(res, &GitWorktreeDesc{
				Path:          worktreePath,
				Size:          size,
				LastAccessAt:  lastAccessAt,
				CacheBasePath: dir,
			})
		}
	}

	return res, nil
}
