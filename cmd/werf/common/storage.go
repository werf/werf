package common

import (
	"context"
	"fmt"

	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/storage/manager"
)

type NewStorageManagerOption func(*NewStorageManagerConfig)

type NewStorageManagerConfig struct {
	ProjectName string

	ContainerBackend container_backend.ContainerBackend
	CmdData          *CmdData

	hostPurge bool

	CleanupDisabled                bool
	GitHistoryBasedCleanupDisabled bool
	SkipMetaCheck                  bool

	// Registry-model requiredness hints (v3). --repo preset satisfies both.
	ImagesRepoRequired bool
	MetaRepoRequired   bool
}

func WithHostPurge() NewStorageManagerOption {
	return func(config *NewStorageManagerConfig) {
		config.hostPurge = true
	}
}

func NewStorageManager(ctx context.Context, c *NewStorageManagerConfig) (*manager.StorageManager, error) {
	return NewStorageManagerWithOptions(ctx, c)
}

func NewStorageManagerWithOptions(ctx context.Context, c *NewStorageManagerConfig, opts ...NewStorageManagerOption) (*manager.StorageManager, error) {
	for _, opt := range opts {
		opt(c)
	}

	if !c.hostPurge {
		if err := ResolveRepos(ctx, c.CmdData, ResolveReposOptions{
			ImagesRepoRequired: c.ImagesRepoRequired,
			MetaRepoRequired:   c.MetaRepoRequired,
		}); err != nil {
			return nil, err
		}
	}

	var stagesStorage storage.PrimaryStagesStorage

	if c.hostPurge {
		stagesStorage = GetLocalStagesStorage(c.ContainerBackend)
	} else {
		var stgErr error
		stagesStorage, stgErr = GetStagesStorage(ctx, c.ContainerBackend, c.CmdData, GetStagesStorageOpts{
			CleanupDisabled:                c.CleanupDisabled,
			GitHistoryBasedCleanupDisabled: c.GitHistoryBasedCleanupDisabled,
			SkipMetaCheck:                  c.SkipMetaCheck,
		})
		if stgErr != nil {
			return nil, stgErr
		}
	}

	if c.hostPurge {
		return &manager.StorageManager{
			ProjectName: c.ProjectName,
			Storages:    manager.NewStorages(manager.NewStoragesConfig{Stages: stagesStorage}),
		}, nil
	}

	storages, err := BuildStorage(ctx, c.ContainerBackend, c.CmdData, stagesStorage)
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
// must have already run against cmdData (NewStorageManagerWithOptions calls
// it before this).
func BuildStorage(ctx context.Context, containerBackend container_backend.ContainerBackend, cmdData *CmdData, stagesStorage storage.PrimaryStagesStorage) (manager.Storages, error) {
	finalImageStorage, err := GetOptionalFinalImageStorage(ctx, containerBackend, cmdData)
	if err != nil {
		return manager.Storages{}, fmt.Errorf("error get final stages storage: %w", err)
	}
	imagesStorage, err := GetOptionalImagesStorage(ctx, containerBackend, cmdData)
	if err != nil {
		return manager.Storages{}, fmt.Errorf("error get images storage: %w", err)
	}
	secondaryStagesStorageList, err := GetSecondaryStagesStorageList(ctx, stagesStorage, containerBackend, cmdData)
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
	metaStorage, err := GetMetaStorage(ctx, containerBackend, cmdData, stagesStorage)
	if err != nil {
		return manager.Storages{}, fmt.Errorf("error get meta stages storage: %w", err)
	}

	return manager.NewStorages(manager.NewStoragesConfig{
		Stages:    stagesStorage,
		Final:     finalImageStorage,
		Images:    imagesStorage,
		Meta:      metaStorage,
		CacheFrom: cacheStagesStorageList,
		CacheTo:   cacheStagesWriteList,
		Secondary: secondaryStagesStorageList,
	}), nil
}
