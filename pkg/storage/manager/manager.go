package manager

import (
	"github.com/werf/werf/pkg/storage"
)

const MaxNumberOfWorkersDefault = 20

type StorageManager struct {
	*stagesStorageManager
}

func NewStorageManager(projectName string, storageLockManager storage.LockManager, stagesStorageCache storage.StagesStorageCache) *StorageManager {
	return &StorageManager{
		stagesStorageManager: newStagesStorageManager(projectName, storageLockManager, stagesStorageCache),
	}
}
