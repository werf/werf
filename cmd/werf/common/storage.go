package common

import (
	"context"
	"fmt"

	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/storage/manager"
)

type NewStorageManagerConfig struct {
	ProjectName string

	ContainerBackend container_backend.ContainerBackend
	CmdData          *CmdData

	// HostPurge builds a local-only manager without resolving any repos.
	HostPurge bool

	CleanupDisabled                bool
	GitHistoryBasedCleanupDisabled bool
	SkipMetaCheck                  bool

	// Registry-model requiredness hints (v3). --repo preset satisfies both.
	ImagesRepoRequired bool
	MetaRepoRequired   bool
}

func NewStorageManager(ctx context.Context, c *NewStorageManagerConfig) (*manager.StorageManager, error) {
	if !c.HostPurge {
		if err := ResolveRepos(ctx, c.CmdData, ResolveReposOptions{
			ImagesRepoRequired: c.ImagesRepoRequired,
			MetaRepoRequired:   c.MetaRepoRequired,
		}); err != nil {
			return nil, err
		}
	}

	if c.HostPurge {
		return &manager.StorageManager{
			ProjectName: c.ProjectName,
			Storages:    manager.NewStorages(manager.NewStoragesConfig{Stages: GetLocalRegistryStorage(c.ContainerBackend)}),
		}, nil
	}

	registryStorage, err := GetStagesStorage(ctx, c.ContainerBackend, c.CmdData, GetStagesStorageOpts{
		CleanupDisabled:                c.CleanupDisabled,
		GitHistoryBasedCleanupDisabled: c.GitHistoryBasedCleanupDisabled,
		SkipMetaCheck:                  c.SkipMetaCheck,
	})
	if err != nil {
		return nil, err
	}

	storages, err := BuildStorage(ctx, c.ContainerBackend, c.CmdData, registryStorage)
	if err != nil {
		return nil, err
	}

	return &manager.StorageManager{
		ProjectName: c.ProjectName,
		Storages:    storages,
	}, nil
}

// BuildStorage resolves every repo/registry in use under the granular
// registry model (--images-repo, --final-repo, --meta-repo, --cache-from,
// --cache-to, secondary) into a single manager.Storages value. ResolveRepos
// must have already run against cmdData (NewStorageManager calls it before
// this).
func BuildStorage(ctx context.Context, containerBackend container_backend.ContainerBackend, cmdData *CmdData, registryStorage storage.RegistryStorage) (manager.Storages, error) {
	finalImagesStorage, err := GetOptionalFinalImagesStorage(ctx, containerBackend, cmdData)
	if err != nil {
		return manager.Storages{}, fmt.Errorf("error get final stages storage: %w", err)
	}
	imagesStorage, err := GetOptionalImagesStorage(ctx, containerBackend, cmdData)
	if err != nil {
		return manager.Storages{}, fmt.Errorf("error get images storage: %w", err)
	}
	secondaryStagesStorageList, err := GetSecondaryStagesStorageList(ctx, registryStorage, containerBackend, cmdData)
	if err != nil {
		return manager.Storages{}, fmt.Errorf("error get secondary stages storage list: %w", err)
	}
	cacheStagesStorageList, err := GetCacheStagesStorageList(ctx, containerBackend, cmdData)
	if err != nil {
		return manager.Storages{}, fmt.Errorf("error get cache storage list: %w", err)
	}
	cacheStagesWriteList, err := GetCacheToStagesStorageList(ctx, containerBackend, cmdData)
	if err != nil {
		return manager.Storages{}, fmt.Errorf("error get cache-to storage list: %w", err)
	}
	metaStorage, err := GetMetaStorage(ctx, containerBackend, cmdData, registryStorage)
	if err != nil {
		return manager.Storages{}, fmt.Errorf("error get meta stages storage: %w", err)
	}

	return manager.NewStorages(manager.NewStoragesConfig{
		Stages:    registryStorage,
		Final:     finalImagesStorage,
		Images:    imagesStorage,
		Meta:      metaStorage,
		CacheFrom: cacheStagesStorageList,
		CacheTo:   cacheStagesWriteList,
		Secondary: secondaryStagesStorageList,
	}), nil
}
