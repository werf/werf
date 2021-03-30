package build

import (
	"context"

	"github.com/werf/logboek"
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
Retry:
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

	if shouldRetry, err := func() (bool, error) {
		defer newConveyor.Terminate(ctx)

		if err := f(newConveyor); manager.ShouldResetStagesStorageCache(err) {
			logboek.Context(ctx).Error().LogF("Will reset stages storage cache due to error: %s\n", err)

			if err := newConveyor.StorageManager.ResetStagesStorageCache(ctx); err != nil {
				return false, err
			}
			return true, nil
		} else {
			return false, err
		}
	}(); err != nil {
		return err
	} else if shouldRetry {
		goto Retry
	}
	return nil
}
