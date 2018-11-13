package git_repo

import (
	"path/filepath"

	"github.com/flant/dapp/pkg/dapp"
)

const GIT_WORKTREE_CACHE_VERSION = "1"

func GetBaseWorkTreeDir() string {
	return filepath.Join(dapp.GetHomeDir(), "git", "worktrees", GIT_WORKTREE_CACHE_VERSION)
}
