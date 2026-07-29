package git_repo

import (
	"path/filepath"

	"github.com/werf/werf/v2/pkg/werf"
)

// Before changing: read the local_cache contract in the package doc of pkg/git_repo/gitdata.
const GitWorktreesCacheVersion = "9"

func GetWorkTreeCacheDir() string {
	return filepath.Join(werf.GetLocalCacheDir(), "git_worktrees", GitWorktreesCacheVersion)
}
