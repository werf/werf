package build

import (
	"context"

	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_backend"
	"github.com/werf/werf/pkg/giterminism_manager"
	"github.com/werf/werf/pkg/storage"
	"github.com/werf/werf/pkg/storage/manager"
)

type ConveyorWithRetryWrapper struct {
	WerfConfig         *config.WerfConfig
	GiterminismManager giterminism_manager.Interface
	ProjectDir         string
	BaseTmpDir         string
	SshAuthSock        string
	ContainerBackend   container_backend.ContainerBackend
	StorageManager     *manager.StorageManager
	StorageLockManager storage.LockManager

	ConveyorOptions ConveyorOptions
}

func NewConveyorWithRetryWrapper(werfConfig *config.WerfConfig, giterminismManager giterminism_manager.Interface, projectDir, baseTmpDir, sshAuthSock string, containerBackend container_backend.ContainerBackend, storageManager *manager.StorageManager, storageLockManager storage.LockManager, opts ConveyorOptions) *ConveyorWithRetryWrapper {
	return &ConveyorWithRetryWrapper{
		WerfConfig:         werfConfig,
		GiterminismManager: giterminismManager,
		ProjectDir:         projectDir,
		BaseTmpDir:         baseTmpDir,
		SshAuthSock:        sshAuthSock,
		ContainerBackend:   containerBackend,
		StorageManager:     storageManager,
		StorageLockManager: storageLockManager,
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
			wrapper.SshAuthSock,
			wrapper.ContainerBackend,
			wrapper.StorageManager,
			wrapper.StorageLockManager,
			wrapper.ConveyorOptions,
		)

		defer newConveyor.Terminate(ctx)

		return f(newConveyor)
	})
}
