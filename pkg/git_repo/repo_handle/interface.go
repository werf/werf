package repo_handle

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// Handle is a solution to get away from the worktree when working with the git repository,
// caching the necessary data from the worktree during initialization,
// and then working exclusively with git objects.
type Handle interface {
	Repository() Repository
	Submodule(submodulePath string) (SubmoduleHandle, error)
	Submodules() []SubmoduleHandle
}

type SubmoduleHandle interface {
	Handle
	Config() *config.Submodule
	Status() *git.SubmoduleStatus
}

type Repository interface {
	CommitObject(h plumbing.Hash) (*object.Commit, error)
	BlobObject(h plumbing.Hash) (*object.Blob, error)
}

func NewHandle(repository *git.Repository) (Handle, error) {
	h := newHandle(repository)

	submoduleHandleList, err := getSubmoduleHandleList(repository)
	if err != nil {
		return nil, err
	}

	h.submoduleHandleList = submoduleHandleList

	return h, nil
}

func getSubmoduleHandleList(parentRepository *git.Repository) ([]SubmoduleHandle, error) {
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
			return nil, err
		}

		submoduleStatus, err := s.Status()
		if err != nil {
			return nil, err
		}

		handle, err := NewHandle(submoduleRepository)
		if err != nil {
			return nil, err
		}

		submoduleHandle := newRepositorySubmoduleHandle(handle, s.Config(), submoduleStatus)
		list = append(list, submoduleHandle)
	}

	return list, nil
}
