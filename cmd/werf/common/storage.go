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
	var stagesStorage storage.PrimaryStagesStorage

	if c.hostPurge {
		stagesStorage = GetLocalStagesStorage(c.ContainerBackend)
	} else {
		var stgErr error
		stagesStorage, stgErr = GetStagesStorage(ctx, c.ContainerBackend, c.CmdData, c.CleanupDisabled, c.GitHistoryBasedCleanupDisabled)
		if stgErr != nil {
			return nil, stgErr
		}
	}

	synchronization, err := GetSynchronization(ctx, c.CmdData, c.ProjectName, stagesStorage)
	if err != nil {
		return nil, fmt.Errorf("error get synchronization: %w", err)
	}

	storageLockManager, err := synchronization.GetStorageLockManager(ctx)
	if err != nil {
		return nil, fmt.Errorf("error get storage lock manager: %w", err)
	}

	if c.hostPurge {
		return &manager.StorageManager{
			ProjectName:                c.ProjectName,
			StagesStorage:              stagesStorage,
			StorageLockManager:         storageLockManager,
			FinalStagesStorage:         nil,
			CacheStagesStorageList:     nil,
			SecondaryStagesStorageList: nil,
		}, nil
	}

	finalStagesStorage, err := GetOptionalFinalStagesStorage(ctx, c.ContainerBackend, c.CmdData)
	if err != nil {
		return nil, fmt.Errorf("error get final stages storage: %w", err)
	}

	secondaryStagesStorageList, err := GetSecondaryStagesStorageList(ctx, stagesStorage, c.ContainerBackend, c.CmdData)
	if err != nil {
		return nil, fmt.Errorf("error get secondary stages storage list: %w", err)
	}
	cacheStagesStorageList, err := GetCacheStagesStorageList(ctx, c.ContainerBackend, c.CmdData)
	if err != nil {
		return nil, fmt.Errorf("error get chache storage list: %w", err)
	}
	return &manager.StorageManager{
		ProjectName:        c.ProjectName,
		StorageLockManager: storageLockManager,

		StagesStorage:              stagesStorage,
		FinalStagesStorage:         finalStagesStorage,
		CacheStagesStorageList:     cacheStagesStorageList,
		SecondaryStagesStorageList: secondaryStagesStorageList,
	}, nil
}
