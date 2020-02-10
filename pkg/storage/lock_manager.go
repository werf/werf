package storage

type LockManager interface {
	LockStage(projectName, signature string) error
	UnlockStage(projectName, signature string) error
	LockStageCache(projectName, signature string) error
	UnlockStageCache(projectName, signature string) error
	LockImage(imageName string) error
	UnlockImage(imageName string) error

	//ReleaseAllStageLocks() error
	//LockAllImagesReadOnly(projectName string) error
	//UnlockAllImages(projectName string) error
}
