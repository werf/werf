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

func GetExistingGitRepos(cacheVersionRoot string) ([]*GitRepoDesc, error) {
	var res []*GitRepoDesc

	if _, err := os.Stat(cacheVersionRoot); os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("error accessing dir %q: %w", cacheVersionRoot, err)
	}

	files, err := ioutil.ReadDir(cacheVersionRoot)
	if err != nil {
		return nil, fmt.Errorf("error reading dir %q: %w", cacheVersionRoot, err)
	}

	for _, finfo := range files {
		repoPath := filepath.Join(cacheVersionRoot, finfo.Name())

		if !finfo.IsDir() {
			if err := os.RemoveAll(repoPath); err != nil {
				return nil, fmt.Errorf("unable to remove %q: %w", repoPath, err)
			}
		}

		size, err := volumeutils.DirSizeBytes(repoPath)
		if err != nil {
			return nil, fmt.Errorf("error getting dir %q size: %w", repoPath, err)
		}

		lastAccessAtPath := filepath.Join(repoPath, "last_access_at")
		lastAccessAt, err := timestamps.ReadTimestampFile(lastAccessAtPath)
		if err != nil {
			return nil, fmt.Errorf("error reading last access timestamp file %q: %w", lastAccessAtPath, err)
		}

		res = append(res, &GitRepoDesc{
			Path:          repoPath,
			Size:          size,
			LastAccessAt:  lastAccessAt,
			CacheBasePath: cacheVersionRoot,
		})
	}

	return res, nil
}
