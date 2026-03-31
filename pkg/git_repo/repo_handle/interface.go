package repo_handle

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/samber/lo"
)

// Handle is a solution to get away from the worktree when working with the git repository,
// caching the necessary data from the worktree during initialization,
// and then working exclusively with git objects.
type Handle interface {
	Repository() *git.Repository
	Submodule(submodulePath string) (SubmoduleHandle, error)
	Submodules() []SubmoduleHandle
	ReadBlobObjectContent(hash plumbing.Hash) ([]byte, error)
	GetCommitTree(hash plumbing.Hash) (TreeHandle, error)
}

type TreeHandle interface {
	Hash() plumbing.Hash
	Entries() []object.TreeEntry
	Tree(path string) (TreeHandle, error)
	FindEntry(path string) (*object.TreeEntry, error)
}

type SubmoduleHandle interface {
	Handle
	Config() *config.Submodule
	Status() *git.SubmoduleStatus
}

type NewHandleOptions struct {
	CommitHash  plumbing.Hash
	WorkTreeDir string
}

func NewHandle(repository *git.Repository, options ...NewHandleOptions) (Handle, error) {
	return newHandleWithSubmodules(repository, &sync.Mutex{}, options...)
}

func newHandleWithSubmodules(repository *git.Repository, mutex *sync.Mutex, options ...NewHandleOptions) (Handle, error) {
	h := newHandle(repository, mutex)

	var opts NewHandleOptions
	if len(options) > 0 {
		opts = options[0]
	}

	var (
		submoduleHandleList []SubmoduleHandle
		err                 error
	)
	if opts.WorkTreeDir != "" {
		submoduleHandleList, err = getSubmoduleHandleListFromCommit(repository, mutex, opts.CommitHash, opts.WorkTreeDir)
	} else {
		submoduleHandleList, err = getSubmoduleHandleListFromWorktree(repository, mutex)
	}
	if err != nil {
		return nil, err
	}

	h.submoduleHandleList = submoduleHandleList

	return h, nil
}

func getSubmoduleHandleListFromWorktree(parentRepository *git.Repository, mutex *sync.Mutex) ([]SubmoduleHandle, error) {
	var list []SubmoduleHandle

	w, err := parentRepository.Worktree()
	if err != nil {
		return nil, err
	}

	ss, err := w.Submodules()
	if err != nil {
		return nil, err
	}

	for _, s := range ss {
		submoduleRepository, err := s.Repository()
		if err != nil {
			return nil, fmt.Errorf("unable to get submodule %q repository: %w", s.Config().Path, err)
		}

		submoduleStatus, err := s.Status()
		if err != nil {
			return nil, fmt.Errorf("unable to get submodule %q status: %w", s.Config().Path, err)
		}

		handle, err := newHandleWithSubmodules(submoduleRepository, mutex)
		if err != nil {
			return nil, err
		}

		submoduleHandle := newSubmoduleHandle(handle, s.Config(), submoduleStatus)
		list = append(list, submoduleHandle)
	}

	return list, nil
}

func getSubmoduleHandleListFromCommit(parentRepository *git.Repository, mutex *sync.Mutex, commitHash plumbing.Hash, workTreeDir string) ([]SubmoduleHandle, error) {
	if commitHash.IsZero() {
		return nil, fmt.Errorf("commit hash is required")
	}
	if workTreeDir == "" {
		return nil, fmt.Errorf("work tree dir is required")
	}

	commit, err := parentRepository.CommitObject(commitHash)
	if err != nil {
		return nil, fmt.Errorf("get commit object %s: %w", commitHash, err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("get commit tree %s: %w", commitHash, err)
	}

	modules, err := readModules(tree)
	if err != nil {
		return nil, err
	}
	if len(modules.Submodules) == 0 {
		return nil, nil
	}

	entries, err := submoduleEntries(tree, modules)
	if err != nil {
		return nil, err
	}

	var list []SubmoduleHandle
	for _, entry := range entries {
		module := entry.module
		submoduleRepository, err := git.PlainOpen(filepath.Join(workTreeDir, module.Path))
		if err != nil {
			return nil, fmt.Errorf("open submodule %q repository: %w", module.Path, err)
		}

		handle, err := newHandleWithSubmodules(submoduleRepository, mutex, NewHandleOptions{
			CommitHash:  entry.hash,
			WorkTreeDir: filepath.Join(workTreeDir, module.Path),
		})
		if err != nil {
			return nil, err
		}

		submoduleStatus := &git.SubmoduleStatus{
			Path:     module.Path,
			Expected: entry.hash,
		}

		submoduleHandle := newSubmoduleHandle(handle, module, submoduleStatus)
		list = append(list, submoduleHandle)
	}

	return list, nil
}

type submoduleEntry struct {
	module *config.Submodule
	hash   plumbing.Hash
}

func readModules(tree *object.Tree) (*config.Modules, error) {
	gitmodulesFile, err := tree.File(".gitmodules")
	if err != nil {
		if errors.Is(err, object.ErrFileNotFound) {
			return config.NewModules(), nil
		}
		return nil, fmt.Errorf("read .gitmodules entry: %w", err)
	}

	gitmodulesContent, err := gitmodulesFile.Contents()
	if err != nil {
		return nil, fmt.Errorf("read .gitmodules: %w", err)
	}

	modules := config.NewModules()
	if err := modules.Unmarshal([]byte(gitmodulesContent)); err != nil {
		return nil, fmt.Errorf("parse .gitmodules: %w", err)
	}

	return modules, nil
}

func submoduleEntries(tree *object.Tree, modules *config.Modules) ([]submoduleEntry, error) {
	var entries []submoduleEntry
	for _, module := range modules.Submodules {
		entry, err := tree.FindEntry(module.Path)
		if err != nil {
			return nil, fmt.Errorf("find submodule entry %q: %w", module.Path, err)
		}
		if entry.Mode != filemode.Submodule {
			return nil, fmt.Errorf("submodule %q has invalid mode %s", module.Path, entry.Mode.String())
		}
		entries = append(entries, submoduleEntry{module: module, hash: entry.Hash})
	}

	return lo.Filter(entries, func(entry submoduleEntry, _ int) bool {
		return entry.module != nil
	}), nil
}
