package build

import (
	"context"

	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/giterminism_manager"
	"github.com/werf/werf/pkg/storage"
	"github.com/werf/werf/pkg/storage/manager"
)

type ConveyorWithRetryWrapper struct {
	WerfConfig          *config.WerfConfig
	GiterminismManager  giterminism_manager.Interface
	ImageNamesToProcess []string
	ProjectDir          string
	BaseTmpDir          string
	SshAuthSock         string
	ContainerRuntime    container_runtime.ContainerRuntime
	StorageManager      *manager.StorageManager
	StorageLockManager  storage.LockManager

	ConveyorOptions ConveyorOptions
}

func NewConveyorWithRetryWrapper(werfConfig *config.WerfConfig, giterminismManager giterminism_manager.Interface, imageNamesToProcess []string, projectDir, baseTmpDir, sshAuthSock string, containerRuntime container_runtime.ContainerRuntime, storageManager *manager.StorageManager, storageLockManager storage.LockManager, opts ConveyorOptions) *ConveyorWithRetryWrapper {
	return &ConveyorWithRetryWrapper{
		WerfConfig:          werfConfig,
		GiterminismManager:  giterminismManager,
		ImageNamesToProcess: imageNamesToProcess,
		ProjectDir:          projectDir,
		BaseTmpDir:          baseTmpDir,
		SshAuthSock:         sshAuthSock,
		ContainerRuntime:    containerRuntime,
		StorageManager:      storageManager,
		StorageLockManager:  storageLockManager,
		ConveyorOptions:     opts,
	}
}

func (wrapper *ConveyorWithRetryWrapper) Terminate() error {
	return nil
}

func (wrapper *ConveyorWithRetryWrapper) WithRetryBlock(ctx context.Context, f func(c *Conveyor) error) error {
	return manager.RetryOnStagesStorageCacheResetError(ctx, wrapper.StorageManager, func() error {
		newConveyor := NewConveyor(
			wrapper.WerfConfig,
			wrapper.GiterminismManager,
			wrapper.ImageNamesToProcess,
			wrapper.ProjectDir,
			wrapper.BaseTmpDir,
			wrapper.SshAuthSock,
			wrapper.ContainerRuntime,
			wrapper.StorageManager,
			wrapper.StorageLockManager,
			wrapper.ConveyorOptions,
		)

		defer newConveyor.Terminate(ctx)

		return f(newConveyor)
	})
}
