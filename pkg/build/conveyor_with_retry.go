package build

import (
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/stages_manager"
	"github.com/werf/werf/pkg/storage"
)

type ConveyorWithRetryWrapper struct {
	WerfConfig          *config.WerfConfig
	ImageNamesToProcess []string
	ProjectDir          string
	BaseTmpDir          string
	SshAuthSock         string
	ContainerRuntime    container_runtime.ContainerRuntime
	StagesManager       *stages_manager.StagesManager
	ImagesRepo          storage.ImagesRepo
	StorageLockManager  storage.LockManager

	ConveyorOptions ConveyorOptions
}

func NewConveyorWithRetryWrapper(werfConfig *config.WerfConfig, imageNamesToProcess []string, projectDir, baseTmpDir, sshAuthSock string, containerRuntime container_runtime.ContainerRuntime, stagesManager *stages_manager.StagesManager, imagesRepo storage.ImagesRepo, storageLockManager storage.LockManager, opts ConveyorOptions) *ConveyorWithRetryWrapper {
	return &ConveyorWithRetryWrapper{
		WerfConfig:          werfConfig,
		ImageNamesToProcess: imageNamesToProcess,
		ProjectDir:          projectDir,
		BaseTmpDir:          baseTmpDir,
		SshAuthSock:         sshAuthSock,
		ContainerRuntime:    containerRuntime,
		StagesManager:       stagesManager,
		ImagesRepo:          imagesRepo,
		StorageLockManager:  storageLockManager,
		ConveyorOptions:     opts,
	}
}

func (wrapper *ConveyorWithRetryWrapper) Terminate() error {
	return nil
}

func (wrapper *ConveyorWithRetryWrapper) WithRetryBlock(f func(c *Conveyor) error) error {
Retry:
	newConveyor := NewConveyor(
		wrapper.WerfConfig,
		wrapper.ImageNamesToProcess,
		wrapper.ProjectDir,
		wrapper.BaseTmpDir,
		wrapper.SshAuthSock,
		wrapper.ContainerRuntime,
		wrapper.StagesManager,
		wrapper.ImagesRepo,
		wrapper.StorageLockManager,
		wrapper.ConveyorOptions,
	)

	if shouldRetry, err := func() (bool, error) {
		defer newConveyor.Terminate()

		if err := f(newConveyor); stages_manager.ShouldResetStagesStorageCache(err) {
			if err := newConveyor.StagesManager.ResetStagesStorageCache(); err != nil {
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
