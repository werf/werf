package git_repo

import (
	"os"
	"path/filepath"
)

const GIT_WORKTREE_CACHE_VERSION = "1"

func GetBaseWorkTreeDir() string {
	return filepath.Join(os.Getenv("HOME"), ".dapp", "git", "worktrees", GIT_WORKTREE_CACHE_VERSION)
}
