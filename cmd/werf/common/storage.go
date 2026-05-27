package common

import (
	"context"
	"fmt"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/storage/manager"
)

type NewStorageManagerConfig struct {
	ProjectName string

	ContainerBackend container_backend.ContainerBackend
	CmdData          *CmdData

	HostPurge bool

	CleanupDisabled                bool
	GitHistoryBasedCleanupDisabled bool
	SkipMetaCheck                  bool
}

func NewStorageManager(ctx context.Context, c *NewStorageManagerConfig) (*manager.StorageManager, error) {
	return NewStorageManagerWithOptions(ctx, c)
}

func NewStorageManagerWithOptions(ctx context.Context, c *NewStorageManagerConfig) (*manager.StorageManager, error) {
	if err := resolveDeprecatedFlags(ctx, c.CmdData); err != nil {
		return nil, err
	}

	if err := validateRepoFlags(c.CmdData); err != nil {
		return nil, err
	}

	var stagesStorage storage.PrimaryStagesStorage

	if c.HostPurge {
		stagesStorage = GetLocalStagesStorage(c.ContainerBackend)
	} else {
		var stgErr error
		if getGranularMode(c.CmdData) {
			if c.CmdData.MetaRepo == nil || c.CmdData.MetaRepo.Address == nil || *c.CmdData.MetaRepo.Address == "" {
				return nil, fmt.Errorf("--meta-repo is required when using granular registry flags (--images-repo, --cache-from, --cache-to)")
			}

			stagesStorage, stgErr = GetMetaRepoStagesStorage(ctx, c.ContainerBackend, c.CmdData, GetStagesStorageOpts{
				CleanupDisabled:                c.CleanupDisabled,
				GitHistoryBasedCleanupDisabled: c.GitHistoryBasedCleanupDisabled,
				SkipMetaCheck:                  c.SkipMetaCheck,
			})
		} else {
			stagesStorage, stgErr = GetStagesStorage(ctx, c.ContainerBackend, c.CmdData, GetStagesStorageOpts{
				CleanupDisabled:                c.CleanupDisabled,
				GitHistoryBasedCleanupDisabled: c.GitHistoryBasedCleanupDisabled,
				SkipMetaCheck:                  c.SkipMetaCheck,
			})
		}
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

	if c.HostPurge {
		sm := &manager.StorageManager{
			ProjectName:          c.ProjectName,
			MetaStorage:          stagesStorage,
			StorageLockManager:   storageLockManager,
			ImagesRepoStorage:    nil,
			CacheToStorageList:   nil,
			CacheFromStorageList: nil,
		}
		printRepositoriesDiagnostic(ctx, sm, c.CmdData)
		return sm, nil
	}

	var imagesRepoStorage storage.StagesStorage
	var secondaryStagesStorageList []storage.StagesStorage
	var cacheToStorageList []storage.StagesStorage
	if getGranularMode(c.CmdData) {
		imagesRepoStorage, err = GetImagesRepoStagesStorage(ctx, c.ContainerBackend, c.CmdData)
		if err != nil {
			return nil, fmt.Errorf("error get images repo storage: %w", err)
		}

		secondaryStagesStorageList, err = getCacheFromStorageList(ctx, c.ContainerBackend, c.CmdData)
		if err != nil {
			return nil, fmt.Errorf("error get cache from storage list: %w", err)
		}
		if stagesStorage.Address() != storage.LocalStorageAddress {
			secondaryStagesStorageList = append([]storage.StagesStorage{storage.NewLocalStagesStorage(c.ContainerBackend)}, secondaryStagesStorageList...)
		}

		cacheToStorageList, err = getCacheToStorageList(ctx, c.ContainerBackend, c.CmdData)
		if err != nil {
			return nil, fmt.Errorf("error get cache storage list: %w", err)
		}
	} else {
		imagesRepoStorage, err = GetOptionalImagesRepoStorage(ctx, c.ContainerBackend, c.CmdData)
		if err != nil {
			return nil, fmt.Errorf("error get images repo storage: %w", err)
		}

		secondaryStagesStorageList, err = GetSecondaryStagesStorageList(ctx, stagesStorage, c.ContainerBackend, c.CmdData)
		if err != nil {
			return nil, fmt.Errorf("error get cache from storage list: %w", err)
		}
		cacheToStorageList, err = GetCacheToStorageList(ctx, c.ContainerBackend, c.CmdData)
		if err != nil {
			return nil, fmt.Errorf("error get cache storage list: %w", err)
		}
	}

	sm := &manager.StorageManager{
		ProjectName:        c.ProjectName,
		StorageLockManager: storageLockManager,

		MetaStorage:          stagesStorage,
		ImagesRepoStorage:    imagesRepoStorage,
		CacheToStorageList:   cacheToStorageList,
		CacheFromStorageList: secondaryStagesStorageList,
	}
	printRepositoriesDiagnostic(ctx, sm, c.CmdData)
	return sm, nil
}

func printRepositoriesDiagnostic(ctx context.Context, sm *manager.StorageManager, cmdData *CmdData) {
	isGranular := getGranularMode(cmdData)
	suffix := ""
	if !isGranular {
		suffix = " (from --repo)"
	}

	logboek.Context(ctx).Default().LogBlock("Using repositories").Do(func() {
		printAddress(ctx, "cache-from", getCacheFromAddresses(sm), suffix)
		printAddress(ctx, "cache-to", getCacheToAddresses(sm), suffix)
		printAddress(ctx, "images-repo", []string{getFinalRepoAddress(sm)}, suffix)
		printAddress(ctx, "meta-repo", []string{sm.MetaStorage.Address()}, suffix)
	})
}

func getCacheFromAddresses(sm *manager.StorageManager) []string {
	addrs := []string{}
	for _, storage := range sm.CacheFromStorageList {
		addrs = append(addrs, storage.Address())
	}
	return addrs
}

func getCacheToAddresses(sm *manager.StorageManager) []string {
	addrs := []string{}
	for _, storage := range sm.CacheToStorageList {
		addrs = append(addrs, storage.Address())
	}
	return addrs
}

func getFinalRepoAddress(sm *manager.StorageManager) string {
	if sm.ImagesRepoStorage != nil {
		return sm.ImagesRepoStorage.Address()
	}
	return "(not set)"
}

func printAddress(ctx context.Context, label string, addrs []string, suffix string) {
	for i, addr := range addrs {
		if i == 0 {
			logboek.Context(ctx).Default().LogF("  %s: %s%s\n", label, addr, suffix)
		} else {
			logboek.Context(ctx).Default().LogF("             %s%s\n", addr, suffix)
		}
	}
}
