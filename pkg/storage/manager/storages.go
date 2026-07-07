package manager

import "github.com/werf/werf/v2/pkg/storage"

// Storages groups every repo/registry werf reads from or writes to under the
// granular registry model (--repo preset or --cache-from/--cache-to/
// --images-repo/--meta-repo/--final-repo). It is a plain data holder: no I/O,
// no mutation — construction and behavior live on StorageManager.
type Storages struct {
	// Stages is the primary raw-stage storage (--repo, or :local by default).
	Stages storage.RegistryStorage
	// Final holds published final images (--final-repo). Nil unless set.
	Final storage.RegistryStorage
	// Images holds content-tag storage — the exclusive home for content-tag
	// resolution and publish (--images-repo, defaults to :local).
	Images storage.RegistryStorage
	// meta holds build/cleanup metadata (--meta-repo). Nil under the --repo
	// preset; use Meta() to get the effective storage with fallback applied.
	meta storage.RegistryStorage
	// CacheFrom is the read list (--cache-from): searched in order when
	// resolving stages.
	CacheFrom []storage.RegistryStorage
	// CacheTo is the write fan-out list (--cache-to): newly built/fetched
	// stages are copied into all of these.
	CacheTo   []storage.RegistryStorage
	Secondary []storage.RegistryStorage
}

// NewStoragesConfig holds every storage needed to construct Storages. meta is
// unexported on Storages itself, so packages outside pkg/storage/manager must
// go through this constructor rather than setting the field directly and
// risking a call site that forgets the --repo-preset fallback.
type NewStoragesConfig struct {
	Stages    storage.RegistryStorage
	Final     storage.RegistryStorage
	Images    storage.RegistryStorage
	Meta      storage.RegistryStorage
	CacheFrom []storage.RegistryStorage
	CacheTo   []storage.RegistryStorage
	Secondary []storage.RegistryStorage
}

func NewStorages(c NewStoragesConfig) Storages {
	return Storages{
		Stages:    c.Stages,
		Final:     c.Final,
		Images:    c.Images,
		meta:      c.Meta,
		CacheFrom: c.CacheFrom,
		CacheTo:   c.CacheTo,
		Secondary: c.Secondary,
	}
}

// Meta returns the effective storage holding build/cleanup metadata. Falls
// back to Stages when no dedicated --meta-repo is set (i.e. the --repo
// preset), preserving co-located behavior bit-for-bit.
func (s *Storages) Meta() storage.RegistryStorage {
	if s.meta != nil {
		return s.meta
	}
	return s.Stages
}

// IsRemoteImagesStorage reports whether the content-tag storage is a remote
// registry, as opposed to :local. Image metadata (managed-image records,
// custom tags, git metadata) is only meaningful for remote registries — a
// local Docker image has no such concept — so this is the single guard that
// decides whether metadata publishing runs at all.
func (s *Storages) IsRemoteImagesStorage() bool {
	_, isLocal := s.Images.(*storage.LocalRegistryStorage)
	return !isLocal
}
