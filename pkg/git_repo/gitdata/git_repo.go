package gitdata

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/volumeutils"
)

type GitRepoDesc struct {
	Path          string
	LastAccessAt  time.Time
	Size          uint64
	CacheBasePath string
}

func (entry *GitRepoDesc) GetPaths() []string {
	return []string{entry.Path}
}

func (entry *GitRepoDesc) GetSize() uint64 {
	return entry.Size
}

func (entry *GitRepoDesc) GetLastAccessAt() time.Time {
	return entry.LastAccessAt
}

func (entry *GitRepoDesc) GetCacheBasePath() string {
	return entry.CacheBasePath
}

// GetGitReposAndRemoveInvalid scans the given cacheVersionRoot directory and returns
// a list of GitRepoDesc for each valid git repository mirror found. It removes
// invalid entries and handles errors appropriately.
//
// The directory structure expected is as follows:
// ├── c447df0d5918decb5d832cb4324e3e2cbe0670eb3fe9301f795be831a9175f47
// │   └── ... (repository files)
// └── ... (other repositories)
//
// Each repo dir is itself a bare full mirror and an independent LRU entry
// with its own last_access_at.
//
// An entry whose size or access marker cannot be read is preserved and left
// out of the result; its error is joined into the returned error so the
// caller can still report a failure after processing the readable entries.
func GetGitReposAndRemoveInvalid(ctx context.Context, cacheVersionRoot string, options ScanOptions) ([]GitDataEntry, error) {
	var res []GitDataEntry
	var errs []error

	// Check if cacheVersionRoot exists and is a directory
	fileStat, err := os.Stat(cacheVersionRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error accessing dir %q: %w", cacheVersionRoot, err)
	}
	if !fileStat.IsDir() {
		logboek.Context(ctx).Warn().LogF("Removing invalid entry %q: not a directory\n", cacheVersionRoot)
		if err := removePath(cacheVersionRoot, options); err != nil {
			return nil, fmt.Errorf("unable to remove %q: %w", cacheVersionRoot, err)
		}
		return nil, nil
	}

	repoDirs, err := ioutil.ReadDir(cacheVersionRoot)
	if err != nil {
		return nil, fmt.Errorf("error reading dir %q: %w", cacheVersionRoot, err)
	}

	for _, repoDirInfo := range repoDirs {
		repoPath := filepath.Join(cacheVersionRoot, repoDirInfo.Name())

		if !repoDirInfo.IsDir() {
			logboek.Context(ctx).Warn().LogF("Removing invalid entry %q: not a directory\n", repoPath)
			if err := removePath(repoPath, options); err != nil {
				return nil, fmt.Errorf("unable to remove %q: %w", repoPath, err)
			}
			continue
		}

		size, err := volumeutils.DirSizeBytes(repoPath)
		if err != nil {
			errs = append(errs, fmt.Errorf("get dir %q size: %w", repoPath, err))
			continue
		}

		lastAccessAtPath := filepath.Join(repoPath, "last_access_at")
		lastAccessAt, err := readLastAccessAt(lastAccessAtPath)
		if err != nil {
			errs = append(errs, fmt.Errorf("read repository access timestamp %q: %w", lastAccessAtPath, err))
			continue
		}

		res = append(res, &GitRepoDesc{
			Path:          repoPath,
			Size:          size,
			LastAccessAt:  lastAccessAt,
			CacheBasePath: cacheVersionRoot,
		})
	}

	return res, errors.Join(errs...)
}
