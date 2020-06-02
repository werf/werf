package git_repo

import (
	"path/filepath"

	"github.com/flant/werf/pkg/werf"
)

const GitWorkTreeCacheVersion = "5"

func GetWorkTreeCacheDir() string {
	return filepath.Join(werf.GetLocalCacheDir(), "git_worktrees", GitWorkTreeCacheVersion)
}
