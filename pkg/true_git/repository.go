package true_git

import (
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/storage/filesystem"
)

func GitOpenWithCustomWorktreeDir(gitDir, worktreeDir string) (*git.Repository, error) {
	worktreeFilesystem := osfs.New(worktreeDir)
	storage := filesystem.NewStorage(osfs.New(gitDir), cache.NewObjectLRUDefault())
	return git.Open(storage, worktreeFilesystem)
}
