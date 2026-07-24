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
// │   ├── full/
// │   │   └── ... (repository files)
// │   ├── shallow/
// │   │   └── ... (repository files)
// │   └── requires_full (optional marker file)
// └── ... (other repositories)
//
// Each full/shallow mirror is an independent LRU entry with its own
// last_access_at. A repo dir with no mirror subdirs is removed entirely,
// including the requires_full marker.
func GetGitReposAndRemoveInvalid(ctx context.Context, cacheVersionRoot string) ([]GitDataEntry, error) {
	var res []GitDataEntry

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
		if err := os.RemoveAll(cacheVersionRoot); err != nil {
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
			if err := os.RemoveAll(repoPath); err != nil {
				return nil, fmt.Errorf("unable to remove %q: %w", repoPath, err)
			}
			continue
		}

		var mirrorEntries []GitDataEntry

		for _, kind := range []string{"full", "shallow"} {
			mirrorPath := filepath.Join(repoPath, kind)

			mirrorStat, err := os.Stat(mirrorPath)
			if os.IsNotExist(err) {
				continue
			} else if err != nil {
				return nil, fmt.Errorf("error accessing dir %q: %w", mirrorPath, err)
			}

			if !mirrorStat.IsDir() {
				logboek.Context(ctx).Warn().LogF("Removing invalid entry %q: not a directory\n", mirrorPath)
				if err := os.RemoveAll(mirrorPath); err != nil {
					return nil, fmt.Errorf("unable to remove %q: %w", mirrorPath, err)
				}
				continue
			}

			size, err := volumeutils.DirSizeBytes(mirrorPath)
			if err != nil {
				return nil, fmt.Errorf("error getting dir %q size: %w", mirrorPath, err)
			}

			lastAccessAtPath := filepath.Join(mirrorPath, "last_access_at")
			lastAccessAt, err := timestamps.ReadTimestampFile(lastAccessAtPath)
			if err != nil {
				logboek.Context(ctx).Warn().LogF("Removing invalid entry %q: error reading last access timestamp file %q: %v\n", mirrorPath, lastAccessAtPath, err)
				if err := os.RemoveAll(mirrorPath); err != nil {
					return nil, fmt.Errorf("unable to remove %q: %w", mirrorPath, err)
				}
				continue
			}

			mirrorEntries = append(mirrorEntries, &GitRepoDesc{
				Path:          mirrorPath,
				Size:          size,
				LastAccessAt:  lastAccessAt,
				CacheBasePath: cacheVersionRoot,
			})
		}

		if len(mirrorEntries) == 0 {
			logboek.Context(ctx).Warn().LogF("Removing invalid entry %q: no valid git mirrors inside\n", repoPath)
			if err := os.RemoveAll(repoPath); err != nil {
				return nil, fmt.Errorf("unable to remove %q: %w", repoPath, err)
			}
			continue
		}

		res = append(res, mirrorEntries...)
	}

	return res, nil
}
