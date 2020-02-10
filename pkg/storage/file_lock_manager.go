package storage

import (
	"fmt"

	"github.com/flant/shluz"
)

type FileLockManager struct {
	stageLocks []string
}

func (lockManager *FileLockManager) LockStage(projectName, signature string) error {
	for _, lockName := range lockManager.stageLocks {
		if lockName == signature {
			return nil
		}
	}

	if err := shluz.Lock(fmt.Sprintf("%s.%s", projectName, signature), shluz.LockOptions{}); err != nil {
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
		if err := shluz.Unlock(fmt.Sprintf("%s.%s", projectName, signature)); err != nil {
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
		if err := shluz.Unlock(lockName); err != nil {
			return err
		}
	}

	return nil
}

func (lockManager *FileLockManager) LockAllImagesReadOnly(projectName string) error {
	lockName := fmt.Sprintf("%s.images", projectName)
	err := shluz.Lock(lockName, shluz.LockOptions{ReadOnly: true})
	if err != nil {
		return fmt.Errorf("shluz lock %s error: %s", lockName, err)
	}
	return nil
}

func (lockManager *FileLockManager) UnlockAllImages(projectName string) error {
	lockName := fmt.Sprintf("%s.images", projectName)
	return shluz.Unlock(lockName)
}

func (lockManager *FileLockManager) LockStageCache(projectName, signature string) error {
	lockName := fmt.Sprintf("%s.%s.cache", projectName, signature)
	if err := shluz.Lock(lockName, shluz.LockOptions{}); err != nil {
		return fmt.Errorf("shluz lock %s error: %s", lockName, err)
	}
	return nil
}

func (lockManager *FileLockManager) UnlockStageCache(projectName, signature string) error {
	lockName := fmt.Sprintf("%s.%s.cache", projectName, signature)
	return shluz.Unlock(lockName)
}

func (lockManager *FileLockManager) LockImage(imageName string) error {
	lockName := fmt.Sprintf("%s.image", imageName)
	if err := shluz.Lock(lockName, shluz.LockOptions{}); err != nil {
		return fmt.Errorf("shluz lock %s error: %s", lockName, err)
	}
	return nil
}

func (lockManager *FileLockManager) UnlockImage(imageName string) error {
	lockName := fmt.Sprintf("%s.image", imageName)
	return shluz.Unlock(lockName)
}
