package git_repo

import (
	"path/filepath"

	"github.com/werf/werf/pkg/werf"
)

const GitWorktreesCacheVersion = "9"

func GetWorkTreeCacheDir() string {
	return filepath.Join(werf.GetLocalCacheDir(), "git_worktrees", GitWorktreesCacheVersion)
}
