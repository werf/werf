package repo_handle

import (
	"fmt"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
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

func NewHandle(repository *git.Repository) (Handle, error) {
	return newHandleWithSubmodules(repository, &sync.Mutex{})
}

func newHandleWithSubmodules(repository *git.Repository, mutex *sync.Mutex) (Handle, error) {
	h := newHandle(repository, mutex)

	submoduleHandleList, err := getSubmoduleHandleList(repository, mutex)
	if err != nil {
		return nil, err
	}

	h.submoduleHandleList = submoduleHandleList

	return h, nil
}

func getSubmoduleHandleList(parentRepository *git.Repository, mutex *sync.Mutex) ([]SubmoduleHandle, error) {
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
