package storage

import (
	"fmt"

	"github.com/flant/lockgate"

	"github.com/flant/werf/pkg/werf"

	"github.com/flant/logboek"
)

type FileLockManager struct {
	stageLocks []string
	// FIXME: create separate locker using lockgate
}

func (lockManager *FileLockManager) LockStage(projectName, signature string) error {
	for _, lockName := range lockManager.stageLocks {
		if lockName == signature {
			return nil
		}
	}

	if _, err := werf.AcquireHostLock(fmt.Sprintf("%s.%s", projectName, signature), lockgate.AcquireOptions{}); err != nil {
		return err
	}

	lockManager.stageLocks = append(lockManager.stageLocks, signature)

	return nil
}

func (lockManager *FileLockManager) UnlockStage(projectName, signature string) error {
	ind := -1
	for i, lockName := range lockManager.stageLocks {
		if lockName == signature {
			ind = i
			break
		}
	}

	if ind >= 0 {
		if err := werf.ReleaseHostLock(fmt.Sprintf("%s.%s", projectName, signature)); err != nil {
			return err
		}
		lockManager.stageLocks = append(lockManager.stageLocks[:ind], lockManager.stageLocks[ind+1:]...)
	}

	return nil
}

func (lockManager *FileLockManager) ReleaseAllStageLocks() error {
	for len(lockManager.stageLocks) > 0 {
		var lockName string
		lockName, lockManager.stageLocks = lockManager.stageLocks[0], lockManager.stageLocks[1:]
		if err := werf.ReleaseHostLock(lockName); err != nil {
			return err
		}
	}

	return nil
}

func (lockManager *FileLockManager) LockAllImagesReadOnly(projectName string) error {
	lockName := fmt.Sprintf("%s.images", projectName)
	_, err := werf.AcquireHostLock(lockName, lockgate.AcquireOptions{Shared: true})
	if err != nil {
		return fmt.Errorf("lock %s error: %s", lockName, err)
	}
	return nil
}

func (lockManager *FileLockManager) UnlockAllImages(projectName string) error {
	lockName := fmt.Sprintf("%s.images", projectName)
	return werf.ReleaseHostLock(lockName)
}

func (lockManager *FileLockManager) LockStageCache(projectName, signature string) error {
	lockName := fmt.Sprintf("%s.%s.cache", projectName, signature)
	if _, err := werf.AcquireHostLock(lockName, lockgate.AcquireOptions{}); err != nil {
		return fmt.Errorf("lock %s error: %s", lockName, err)
	}
	return nil
}

func (lockManager *FileLockManager) UnlockStageCache(projectName, signature string) error {
	lockName := fmt.Sprintf("%s.%s.cache", projectName, signature)
	return werf.ReleaseHostLock(lockName)
}

func (lockManager *FileLockManager) LockImage(imageName string) error {
	lockName := fmt.Sprintf("%s.image", imageName)
	if _, err := werf.AcquireHostLock(lockName, lockgate.AcquireOptions{}); err != nil {
		return fmt.Errorf("lock %s error: %s", lockName, err)
	}
	return nil
}

func (lockManager *FileLockManager) UnlockImage(imageName string) error {
	lockName := fmt.Sprintf("%s.image", imageName)
	return werf.ReleaseHostLock(lockName)
}

func (lockManager *FileLockManager) LockStagesAndImages(projectName string, opts LockStagesAndImagesOptions) error {
	lockName := fmt.Sprintf("%s.stages_and_images", projectName)
	logboek.Debug.LogF("-- FileLockManager.LockStagesAndImages(%s, %#v) lockName=%q\n", projectName, opts, lockName)
	if _, err := werf.AcquireHostLock(lockName, lockgate.AcquireOptions{Shared: opts.GetOrCreateImagesOnly}); err != nil {
		return fmt.Errorf("lock %s error: %s", lockName, err)
	}
	return nil
}

func (lockManager *FileLockManager) UnlockStagesAndImages(projectName string) error {
	lockName := fmt.Sprintf("%s.stages_and_images", projectName)
	logboek.Debug.LogF("-- FileLockManager.UnlockStagesAndImages(%s) lockName=%q\n", projectName, lockName)
	if err := werf.ReleaseHostLock(lockName); err != nil {
		return fmt.Errorf("unlock %s error: %s", lockName, err)
	}
	return nil
}
