package main

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/storage/filesystem"
)

func GitOpenWithCustomWorktreeDir(gitDir, worktreeDir string) (*git.Repository, error) {
	worktreeFilesystem := osfs.New(worktreeDir)
	storage := filesystem.NewStorage(osfs.New(gitDir), cache.NewObjectLRUDefault())
	return git.Open(storage, worktreeFilesystem)
}

func newHash(s string) (plumbing.Hash, error) {
	var h plumbing.Hash

	b, err := hex.DecodeString(s)
	if err != nil {
		return h, err
	}

	copy(h[:], b)
	return h, nil
}

func do(gitDir, workTreeDir, commit string) error {
	if repository, err := GitOpenWithCustomWorktreeDir(gitDir, workTreeDir); err != nil {
		return fmt.Errorf("plain open %s failed: %s", workTreeDir, err)
	} else {
		hash, err := newHash(commit)
		if err != nil {
			return fmt.Errorf("bad commit %q hash: %s", commit, err)
		}

		commit, err := repository.CommitObject(hash)
		if err != nil {
			return fmt.Errorf("commit object fetch failed: %s", err)
		}

		fmt.Printf("-- Commit -> %v\n", commit)
	}

	return nil
}

func main() {
	if err := do(os.Args[1], os.Args[2], os.Args[3]); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
}
