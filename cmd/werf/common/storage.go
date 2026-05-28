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

	isLocalMode := c.HostPurge
	if !c.HostPurge && !getGranularMode(c.CmdData) && c.CmdData.Repo != nil && c.CmdData.Repo.Address != nil && *c.CmdData.Repo.Address == storage.LocalStorageAddress {
		isLocalMode = true
	}

	if isLocalMode {
		localStg := GetLocalStagesStorage(c.ContainerBackend)

		synchronization, err := GetSynchronization(ctx, c.CmdData, c.ProjectName, nil)
		if err != nil {
			return nil, fmt.Errorf("error get synchronization: %w", err)
		}

		storageLockManager, err := synchronization.GetStorageLockManager(ctx)
		if err != nil {
			return nil, fmt.Errorf("error get storage lock manager: %w", err)
		}

		sm := &manager.StorageManager{
			ProjectName:        c.ProjectName,
			MetaStorage:        nil,
			StorageLockManager: storageLockManager,
			ImagesRepoStorages: nil,
			CacheWriters:       []storage.StageWriter{localStg},
			CacheReaders:       []storage.StageReader{localStg},
		}
		warnIfCacheListsEmpty(ctx, sm)
		printRepositoriesDiagnostic(ctx, sm, c.CmdData)
		return sm, nil
	}

	var stagesStorage storage.CacheAndMetaStorage
	var stgErr error
	if getGranularMode(c.CmdData) {
		if c.CmdData.MetaRepo == nil || c.CmdData.MetaRepo.Address == nil || *c.CmdData.MetaRepo.Address == "" {
			return nil, fmt.Errorf("--meta-repo is required when using granular registry flags (--images-repo, --cache-from, --cache-to)")
		}

		stagesStorage, stgErr = GetMetaRepoStagesStorage(ctx, c.ContainerBackend, c.CmdData, GetMetaStorageOpts{
			CleanupDisabled:                c.CleanupDisabled,
			GitHistoryBasedCleanupDisabled: c.GitHistoryBasedCleanupDisabled,
			SkipMetaCheck:                  c.SkipMetaCheck,
		})
	} else {
		stagesStorage, stgErr = GetMetaStorage(ctx, c.ContainerBackend, c.CmdData, GetMetaStorageOpts{
			CleanupDisabled:                c.CleanupDisabled,
			GitHistoryBasedCleanupDisabled: c.GitHistoryBasedCleanupDisabled,
			SkipMetaCheck:                  c.SkipMetaCheck,
		})
	}
	if stgErr != nil {
		return nil, stgErr
	}

	synchronization, err := GetSynchronization(ctx, c.CmdData, c.ProjectName, stagesStorage)
	if err != nil {
		return nil, fmt.Errorf("error get synchronization: %w", err)
	}

	storageLockManager, err := synchronization.GetStorageLockManager(ctx)
	if err != nil {
		return nil, fmt.Errorf("error get storage lock manager: %w", err)
	}

	var imagesRepoStorages []storage.ImagesRepoStorage
	var cacheFromList []storage.StageReader
	var cacheToList []storage.StageWriter
	if getGranularMode(c.CmdData) {
		imagesRepoStorages, err = GetImagesRepoStorages(ctx, c.ContainerBackend, c.CmdData)
		if err != nil {
			return nil, fmt.Errorf("error get images repo storage: %w", err)
		}

		cacheFromList, err = getCacheFromStorageList(ctx, c.ContainerBackend, c.CmdData)
		if err != nil {
			return nil, fmt.Errorf("error get cache from storage list: %w", err)
		}
		cacheFromList = append([]storage.StageReader{storage.NewLocalStagesStorage(c.ContainerBackend)}, cacheFromList...)

		cacheToList, err = getCacheToStorageList(ctx, c.ContainerBackend, c.CmdData)
		if err != nil {
			return nil, fmt.Errorf("error get cache storage list: %w", err)
		}
	} else {
		imagesRepoStorages, err = GetOptionalImagesRepoStorages(ctx, c.ContainerBackend, c.CmdData)
		if err != nil {
			return nil, fmt.Errorf("error get images repo storage: %w", err)
		}

		cacheFromList, err = getCacheFromStorageList(ctx, c.ContainerBackend, c.CmdData)
		if err != nil {
			return nil, fmt.Errorf("error get cache from storage list: %w", err)
		}
		cacheFromList = append([]storage.StageReader{storage.NewLocalStagesStorage(c.ContainerBackend)}, cacheFromList...)

		cacheToList, err = GetCacheToStorageList(ctx, c.ContainerBackend, c.CmdData)
		if err != nil {
			return nil, fmt.Errorf("error get cache storage list: %w", err)
		}
	}

	sm := &manager.StorageManager{
		ProjectName:        c.ProjectName,
		StorageLockManager: storageLockManager,

		MetaStorage:        stagesStorage,
		ImagesRepoStorages: imagesRepoStorages,
		CacheWriters:       append([]storage.StageWriter{stagesStorage}, cacheToList...),
		CacheReaders:       append([]storage.StageReader{stagesStorage}, cacheFromList...),
	}
	warnIfCacheListsEmpty(ctx, sm)
	printRepositoriesDiagnostic(ctx, sm, c.CmdData)
	return sm, nil
}

func warnIfCacheListsEmpty(ctx context.Context, sm *manager.StorageManager) {
	if len(sm.CacheReaders) == 0 {
		logboek.Context(ctx).Warn().LogF("WARNING: cache readers are empty: building without cache, all stages will be built from scratch\n")
	}
	if len(sm.CacheWriters) == 0 {
		logboek.Context(ctx).Warn().LogF("WARNING: cache writers are empty: built stages will not be pushed to any remote cache, subsequent builds will start from scratch\n")
	}
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
		printAddress(ctx, "images-repo", getImagesRepoAddresses(sm), suffix)
		metaAddrs := []string{}
		if sm.MetaStorage != nil {
			metaAddrs = []string{sm.MetaStorage.Address()}
		}
		printAddress(ctx, "meta-repo", metaAddrs, suffix)
	})
}

func getCacheFromAddresses(sm *manager.StorageManager) []string {
	addrs := []string{}
	for _, s := range sm.CacheReaders {
		addrs = append(addrs, s.Address())
	}
	return addrs
}

func getCacheToAddresses(sm *manager.StorageManager) []string {
	addrs := []string{}
	for _, s := range sm.CacheWriters {
		addrs = append(addrs, s.Address())
	}
	return addrs
}

func getImagesRepoAddresses(sm *manager.StorageManager) []string {
	addrs := []string{}
	for _, s := range sm.ImagesRepoStorages {
		addrs = append(addrs, s.Address())
	}
	if len(addrs) == 0 {
		addrs = []string{"(not set)"}
	}
	return addrs
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
