package common

import (
	"context"
	"fmt"

	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/storage/manager"
	"github.com/werf/werf/v2/pkg/storage/synchronization/lock_manager"
	"github.com/werf/werf/v2/pkg/util/option"
	"github.com/werf/werf/v2/pkg/werf"
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
}

func WithHostPurge() NewStorageManagerOption {
	return func(config *NewStorageManagerConfig) {
		config.hostPurge = true
	}
}

func NewStorageManager(ctx context.Context, c *NewStorageManagerConfig) (*manager.StorageManager, error) {
	return NewStorageManagerWithOptions(ctx, c)
}

func getStorageLockManager(ctx context.Context, c *NewStorageManagerConfig, stagesStorage storage.PrimaryStagesStorage) (lock_manager.Interface, error) {
	if isImagesReadOnly(c.CmdData) {
		var syncAddress string
		if c.CmdData != nil {
			syncAddress = option.PtrValueOrDefault(c.CmdData.Synchronization, "")
		}
		if err := validateSynchronizationParams(syncAddress, stagesStorage.Address()); err != nil {
			return nil, err
		}
		return lock_manager.NewGeneric(werf.HostLocker().Locker()), nil
	}

	synchronization, err := GetSynchronization(ctx, c.CmdData, c.ProjectName, stagesStorage)
	if err != nil {
		return nil, fmt.Errorf("error get synchronization: %w", err)
	}

	storageLockManager, err := synchronization.GetStorageLockManager(ctx)
	if err != nil {
		return nil, fmt.Errorf("error get storage lock manager: %w", err)
	}
	return storageLockManager, nil
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
		stagesStorage, stgErr = GetStagesStorage(ctx, c.ContainerBackend, c.CmdData, GetStagesStorageOpts{
			CleanupDisabled:                c.CleanupDisabled,
			GitHistoryBasedCleanupDisabled: c.GitHistoryBasedCleanupDisabled,
			SkipMetaCheck:                  c.SkipMetaCheck,
		})
		if stgErr != nil {
			return nil, stgErr
		}
	}

	storageLockManager, err := getStorageLockManager(ctx, c, stagesStorage)
	if err != nil {
		return nil, err
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
	cacheStagesStorageList, err := GetCacheStagesStorageList(ctx, stagesStorage, c.ContainerBackend, c.CmdData)
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
