package manager

import (
	"github.com/werf/werf/pkg/storage"
)

type StorageManager struct {
	*StagesStorageManager
}

func NewStorageManager(projectName string, stagesStorage storage.StagesStorage, secondaryStagesStorageList []storage.StagesStorage, storageLockManager storage.LockManager, stagesStorageCache storage.StagesStorageCache) *StorageManager {
	return &StorageManager{
		StagesStorageManager: newStagesStorageManager(projectName, stagesStorage, secondaryStagesStorageList, storageLockManager, stagesStorageCache),
	}
}

type baseManager struct {
	parallel           bool
	parallelTasksLimit int
}

func (m *baseManager) EnableParallel(parallelTasksLimit int) {
	m.parallel = true
	m.parallelTasksLimit = parallelTasksLimit
}

func (m *baseManager) MaxNumberOfWorkers() int {
	if m.parallel && m.parallelTasksLimit > 0 {
		return m.parallelTasksLimit
	}

	return 1
}
