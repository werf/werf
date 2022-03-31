package repo_handle

import (
	"fmt"
	"sync"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type treeHandle struct {
	tree  *object.Tree
	mutex *sync.Mutex
}

func newTreeHandle(tree *object.Tree, mutex *sync.Mutex) *treeHandle {
	return &treeHandle{tree: tree, mutex: mutex}
}

func (h *treeHandle) Hash() plumbing.Hash {
	return h.tree.Hash
}

func (h *treeHandle) Tree(path string) (TreeHandle, error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	tree, err := h.tree.Tree(path)
	if err != nil {
		return nil, fmt.Errorf("unable to get tree %q: %w", path, err)
	}

	treeHandle := newTreeHandle(tree, h.mutex)
	return treeHandle, nil
}

func (h *treeHandle) FindEntry(path string) (*object.TreeEntry, error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	return h.tree.FindEntry(path)
}

func (h *treeHandle) Entries() []object.TreeEntry {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	return h.tree.Entries
}
