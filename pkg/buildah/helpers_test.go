package buildah

import "go.podman.io/storage"

type imageLookupCountingStore struct {
	storage.Store
	imageLookups int
}

var _ storage.Store = (*imageLookupCountingStore)(nil)

func (store *imageLookupCountingStore) Image(id string) (*storage.Image, error) {
	store.imageLookups++
	return store.Store.Image(id)
}
