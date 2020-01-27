package stages_storage

type LockManager interface {
	LockStage(projectName, signature string) error
	//TryLockStage(projectName ,signature string) (bool, error)
	UnlockStage(projectName, signature string) error
	ReleaseAllStageLocks() error
	LockAllImagesReadOnly(projectName string) error
	UnlockAllImages(projectName string) error
}
