package common

import (
	"context"
	"fmt"

	"github.com/werf/werf/v3/pkg/container_backend"
	"github.com/werf/werf/v3/pkg/storage"
	"github.com/werf/werf/v3/pkg/storage/manager"
	"github.com/werf/werf/v3/pkg/storage/synchronization/lock_manager"
	"github.com/werf/werf/v3/pkg/util/option"
	"github.com/werf/werf/v3/pkg/werf"
)

type NewStorageManagerOption func(*NewStorageManagerConfig)

type NewStorageManagerConfig struct {
	ProjectName string

	ContainerBackend container_backend.ContainerBackend
	CmdData          *CmdData

	hostPurge bool

	skipMetaRepoSafeguard bool

	CleanupDisabled                bool
	GitHistoryBasedCleanupDisabled bool
	SkipMetaCheck                  bool
}

func WithHostPurge() NewStorageManagerOption {
	return func(config *NewStorageManagerConfig) {
		config.hostPurge = true
	}
}

// WithSkipMetaRepoSafeguard disables meta-repo marker validation and planting.
// Only the 'werf meta-repo' commands may use it: they legitimately operate on a
// repo state that the safeguard would otherwise reject.
func WithSkipMetaRepoSafeguard() NewStorageManagerOption {
	return func(config *NewStorageManagerConfig) {
		config.skipMetaRepoSafeguard = true
	}
}

func NewStorageManager(ctx context.Context, c *NewStorageManagerConfig) (*manager.StorageManager, error) {
	return NewStorageManagerWithOptions(ctx, c)
}

// getStorageLockManager initializes synchronization for the storage manager.
// Synchronization init writes to the primary repo (it registers a client-id
// record there), so commands that only check for already built images get
// host-local locks instead: they publish nothing, so there is nothing to
// synchronize with other hosts. The configured synchronization address is still
// validated, just without contacting any server.
func getStorageLockManager(ctx context.Context, c *NewStorageManagerConfig, stagesStorage storage.PrimaryStagesStorage) (lock_manager.Interface, error) {
	if IsImagesReadOnly(c.CmdData) {
		var syncAddress string
		if c.CmdData != nil {
			syncAddress = option.PtrValueOrDefault(c.CmdData.Synchronization, "")
		}
		if err := ValidateSynchronizationParams(syncAddress, stagesStorage.Address()); err != nil {
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
		skipMetaCheck := c.SkipMetaCheck
		if c.CmdData != nil && c.CmdData.MetaRepo != nil && c.CmdData.MetaRepo.Address != nil && *c.CmdData.MetaRepo.Address != "" {
			skipMetaCheck = true
		}

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

	if !c.skipMetaRepoSafeguard {
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
