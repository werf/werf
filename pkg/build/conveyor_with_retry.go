package build

import (
	"fmt"

	"github.com/flant/werf/pkg/config"
	"github.com/flant/werf/pkg/container_runtime"
	"github.com/flant/werf/pkg/stages_manager"
	"github.com/flant/werf/pkg/storage"
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

	conveyorsToTerminate []*Conveyor
}

func NewConveyorWithRetryWrapper(werfConfig *config.WerfConfig, imageNamesToProcess []string, projectDir, baseTmpDir, sshAuthSock string, containerRuntime container_runtime.ContainerRuntime, stagesManager *stages_manager.StagesManager, imagesRepo storage.ImagesRepo, storageLockManager storage.LockManager) *ConveyorWithRetryWrapper {
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
	}
}

func (wrapper *ConveyorWithRetryWrapper) Terminate() error {
	var terminateErrors []error
	for _, conveyorToTerminate := range wrapper.conveyorsToTerminate {
		if err := conveyorToTerminate.Terminate(); err != nil {
			terminateErrors = append(terminateErrors, err)
		}
	}

	if len(terminateErrors) > 0 {
		return fmt.Errorf("there were errors during conveyors termination")
	}
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
	)
	wrapper.conveyorsToTerminate = append(wrapper.conveyorsToTerminate, newConveyor)

	if err := f(newConveyor); stages_manager.ShouldResetStagesStorageCache(err) {
		if err := newConveyor.StagesManager.ResetStagesStorageCache(); err != nil {
			return err
		}
		goto Retry
	} else if err != nil {
		return err
	}

	return nil
}
