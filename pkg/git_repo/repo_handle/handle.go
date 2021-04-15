package repo_handle

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
)

type handle struct {
	repository          *git.Repository
	submoduleHandleList []SubmoduleHandle
}

func newHandle(repository *git.Repository) *handle {
	return &handle{
		repository: repository,
	}
}

func (h *handle) Repository() Repository {
	return h.repository
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

type submoduleHandle struct {
	Handle
	config *config.Submodule
	status *git.SubmoduleStatus
}

func newRepositorySubmoduleHandle(handle Handle, config *config.Submodule, status *git.SubmoduleStatus) *submoduleHandle {
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
