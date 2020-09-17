package manager

import (
	"github.com/werf/werf/pkg/storage"
)

const MaxNumberOfWorkersDefault = 20

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
	parallel bool
}

func (m *baseManager) EnableParallel() {
	m.parallel = true
}

func (m *baseManager) MaxNumberOfWorkers() int {
	if m.parallel {
		return MaxNumberOfWorkersDefault
	}

	return 1
}
