package repo_handle

import (
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
)

type handle struct {
	repository          *git.Repository
	submoduleHandleList []SubmoduleHandle

	mutex *sync.Mutex
}

func newHandle(repository *git.Repository, mutex *sync.Mutex) *handle {
	return &handle{
		repository: repository,
		mutex:      mutex,
	}
}

func (h *handle) ReadBlobObjectContent(hash plumbing.Hash) ([]byte, error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	obj, err := h.repository.BlobObject(hash)
	if err != nil {
		return nil, fmt.Errorf("unable to get blob %q object: %w", hash, err)
	}

	f, err := obj.Reader()
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("unable to read blob %q content: %w", hash, err)
	}

	return data, nil
}

func (h *handle) GetCommitTree(hash plumbing.Hash) (TreeHandle, error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	commit, err := h.repository.CommitObject(hash)
	if err != nil {
		return nil, fmt.Errorf("unable to get commit %q object: %w", hash, err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("unable to get commit %q tree: %w", hash, err)
	}

	treeHandle := newTreeHandle(tree, h.mutex)
	return treeHandle, nil
}

func (h *handle) Submodule(submodulePath string) (SubmoduleHandle, error) {
	for _, s := range h.submoduleHandleList {
		if s.Config().Path == submodulePath {
			return s, nil
		}
	}

	return nil, fmt.Errorf("submodule %q not found", submodulePath)
}

func (h *handle) Submodules() []SubmoduleHandle {
	return h.submoduleHandleList
}

func (h *handle) Repository() *git.Repository {
	return h.repository
}

type submoduleHandle struct {
	Handle
	config *config.Submodule
	status *git.SubmoduleStatus
}

func newSubmoduleHandle(handle Handle, config *config.Submodule, status *git.SubmoduleStatus) *submoduleHandle {
	return &submoduleHandle{
		Handle: handle,
		config: config,
		status: status,
	}
}

func (h *submoduleHandle) Config() *config.Submodule {
	return h.config
}

func (h *submoduleHandle) Status() *git.SubmoduleStatus {
	return h.status
}
