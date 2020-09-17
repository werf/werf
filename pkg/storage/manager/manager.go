package manager

import (
	"github.com/werf/werf/pkg/storage"
)

type StorageManager struct {
	*ImagesRepoManager
	*StagesStorageManager
}

func NewStorageManager(projectName string, storageLockManager storage.LockManager, stagesStorageCache storage.StagesStorageCache) *StorageManager {
	return &StorageManager{
		StagesStorageManager: newStagesStorageManager(projectName, storageLockManager, stagesStorageCache),
	}
}

func (m *StorageManager) SetImageRepo(imagesRepo storage.ImagesRepo) {
	m.ImagesRepoManager = newImagesRepoManager(imagesRepo)
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
