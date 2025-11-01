package build

import (
	"context"

	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/giterminism_manager"
	"github.com/werf/werf/v2/pkg/storage/manager"
	"github.com/werf/werf/v2/pkg/storage/synchronization/lock_manager"
)

type ConveyorWithRetryWrapper struct {
	WerfConfig         *config.WerfConfig
	GiterminismManager giterminism_manager.Interface
	ProjectDir         string
	BaseTmpDir         string
	ContainerBackend   container_backend.ContainerBackend
	StorageManager     *manager.StorageManager
	StorageLockManager lock_manager.Interface

	ConveyorOptions ConveyorOptions
}

func NewConveyorWithRetryWrapper(werfConfig *config.WerfConfig, giterminismManager giterminism_manager.Interface, projectDir, baseTmpDir string, containerBackend container_backend.ContainerBackend, storageManager *manager.StorageManager, storageLockManager lock_manager.Interface, opts ConveyorOptions) *ConveyorWithRetryWrapper {
	return &ConveyorWithRetryWrapper{
		WerfConfig:         werfConfig,
		GiterminismManager: giterminismManager,
		ProjectDir:         projectDir,
		BaseTmpDir:         baseTmpDir,
		ContainerBackend:   containerBackend,
		StorageManager:     storageManager,
		StorageLockManager: storageLockManager, // TODO: refactor
		ConveyorOptions:    opts,
	}
}

func (wrapper *ConveyorWithRetryWrapper) Terminate() error {
	return nil
}

func (wrapper *ConveyorWithRetryWrapper) WithRetryBlock(ctx context.Context, f func(c *Conveyor) error) error {
	return manager.RetryOnUnexpectedStagesStorageState(ctx, wrapper.StorageManager, func() error {
		newConveyor := NewConveyor(
			wrapper.WerfConfig,
			wrapper.GiterminismManager,
			wrapper.ProjectDir,
			wrapper.BaseTmpDir,
			wrapper.ContainerBackend,
			wrapper.StorageManager,
			wrapper.StorageLockManager,
			wrapper.ConveyorOptions,
		)

		defer newConveyor.Terminate(ctx)

		return f(newConveyor)
	})
}
