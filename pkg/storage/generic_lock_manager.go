package storage

import (
	"fmt"

	"github.com/flant/lockgate"
	"github.com/flant/logboek"
	"github.com/flant/werf/pkg/werf"
)

func NewGenericLockManager(locker lockgate.Locker) *GenericLockManager {
	return &GenericLockManager{Locker: locker}
}

type GenericLockManager struct {
	// Single Locker for all projects
	Locker lockgate.Locker
}

type LockStagesAndImagesOptions struct {
	GetOrCreateImagesOnly bool
}

func (manager *GenericLockManager) LockStage(projectName, signature string) (LockHandle, error) {
	_, lock, err := manager.Locker.Acquire(genericStageLockName(projectName, signature), werf.SetupLockerDefaultOptions(lockgate.AcquireOptions{}))
	return LockHandle{LockgateHandle: lock, ProjectName: projectName}, err
}

func (manager *GenericLockManager) LockStageCache(projectName, signature string) (LockHandle, error) {
	_, lock, err := manager.Locker.Acquire(genericStageCacheLockName(projectName, signature), werf.SetupLockerDefaultOptions(lockgate.AcquireOptions{}))
	return LockHandle{LockgateHandle: lock, ProjectName: projectName}, err
}

func (manager *GenericLockManager) LockImage(projectName, imageName string) (LockHandle, error) {
	_, lock, err := manager.Locker.Acquire(genericImageLockName(imageName), werf.SetupLockerDefaultOptions(lockgate.AcquireOptions{}))
	return LockHandle{LockgateHandle: lock, ProjectName: projectName}, err
}

func (manager *GenericLockManager) LockStagesAndImages(projectName string, opts LockStagesAndImagesOptions) (LockHandle, error) {
	_, lock, err := manager.Locker.Acquire(genericStagesAndImagesLockName(projectName), werf.SetupLockerDefaultOptions(lockgate.AcquireOptions{Shared: opts.GetOrCreateImagesOnly}))
	return LockHandle{LockgateHandle: lock, ProjectName: projectName}, err
}

func (manager *GenericLockManager) Unlock(lock LockHandle) error {
	err := manager.Locker.Release(lock.LockgateHandle)
	if err != nil {
		logboek.ErrF("ERROR: unable to release lock for %q: %s", lock.LockgateHandle.LockName, err)
	}
	return err
}

func genericStageLockName(projectName, signature string) string {
	return fmt.Sprintf("%s.%s", projectName, signature)
}

func genericStageCacheLockName(projectName, signature string) string {
	return fmt.Sprintf("%s.%s.cache", projectName, signature)
}

func genericImageLockName(imageName string) string {
	return fmt.Sprintf("%s.image", imageName)
}

func genericStagesAndImagesLockName(projectName string) string {
	return fmt.Sprintf("%s.stages_and_images", projectName)
}
