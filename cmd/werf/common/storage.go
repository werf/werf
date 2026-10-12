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

	// SkipMetaRepoSafeguard is reserved for migration and detach commands.
	SkipMetaRepoSafeguard bool

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
		separateMetaRepo, err := hasSeparateMetaRepo(ctx, c.CmdData)
		if err != nil {
			return nil, err
		}
		skipMetaCheck := c.SkipMetaCheck || separateMetaRepo

		var stgErr error
		stagesStorage, stgErr = GetStagesStorage(ctx, c.ContainerBackend, c.CmdData, GetStagesStorageOpts{
			CleanupDisabled:                c.CleanupDisabled,
			GitHistoryBasedCleanupDisabled: c.GitHistoryBasedCleanupDisabled,
			SkipMetaCheck:                  skipMetaCheck,
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
			MetaStorage:                stagesStorage,
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

	metaStorage, err := GetOptionalMetaStorage(ctx, c.ContainerBackend, c.CmdData, stagesStorage, c.CleanupDisabled)
	if err != nil {
		return nil, fmt.Errorf("error get meta storage: %w", err)
	}

	if !c.SkipMetaRepoSafeguard {
		metaStorage, err = storage.SetupMetaRepoSafeguard(ctx, c.ProjectName, stagesStorage, metaStorage, c.CleanupDisabled)
		if err != nil {
			return nil, err
		}
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
		MetaStorage:                metaStorage,
		FinalStagesStorage:         finalStagesStorage,
		CacheStagesStorageList:     cacheStagesStorageList,
		SecondaryStagesStorageList: secondaryStagesStorageList,
	}, nil
}

func hasSeparateMetaRepo(ctx context.Context, cmdData *CmdData) (bool, error) {
	if cmdData == nil || cmdData.MetaRepo == nil || cmdData.MetaRepo.Address == nil || *cmdData.MetaRepo.Address == "" {
		return false, nil
	}
	if cmdData.Repo == nil || cmdData.Repo.Address == nil {
		return false, fmt.Errorf("--meta-repo requires --repo to be set")
	}
	if *cmdData.Repo.Address == "" || *cmdData.Repo.Address == storage.LocalStorageAddress {
		return false, nil
	}
	repoAddress, err := storage.CanonicalRepoAddress(ctx, *cmdData.Repo.Address)
	if err != nil {
		return false, fmt.Errorf("parse --repo: %w", err)
	}
	metaAddress, err := storage.CanonicalRepoAddress(ctx, *cmdData.MetaRepo.Address)
	if err != nil {
		return false, fmt.Errorf("parse --meta-repo: %w", err)
	}
	return repoAddress != metaAddress, nil
}
