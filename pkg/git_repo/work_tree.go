package git_repo

import (
	"path/filepath"

	"github.com/flant/werf/pkg/werf"
)

const GIT_WORKTREE_CACHE_VERSION = "2"

func GetBaseWorkTreeDir() string {
	return filepath.Join(werf.GetLocalCacheDir(), "git_worktrees", GIT_WORKTREE_CACHE_VERSION)
}
