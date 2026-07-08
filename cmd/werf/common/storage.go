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
			ProjectName:                c.ProjectName,
			StagesStorage:              stagesStorage,
			MetaStorage:                stagesStorage,
			FinalStagesStorage:         nil,
			CacheStagesStorageList:     nil,
			SecondaryStagesStorageList: nil,
		}, nil
	}

	finalStagesStorage, err := GetOptionalFinalStagesStorage(ctx, c.ContainerBackend, c.CmdData)
	if err != nil {
		return nil, fmt.Errorf("error get final stages storage: %w", err)
	}

	metaStorage, err := GetOptionalMetaStorage(ctx, c.ContainerBackend, c.CmdData, stagesStorage)
	if err != nil {
		return nil, fmt.Errorf("error get meta storage: %w", err)
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
		ProjectName: c.ProjectName,

		StagesStorage:              stagesStorage,
		MetaStorage:                metaStorage,
		FinalStagesStorage:         finalStagesStorage,
		CacheStagesStorageList:     cacheStagesStorageList,
		SecondaryStagesStorageList: secondaryStagesStorageList,
	}, nil
}
