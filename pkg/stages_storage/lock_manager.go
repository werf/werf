package stages_storage

type LockManager interface {
	LockStage(projectName, signature string) error
	UnlockStage(projectName, signature string) error
	LockStageCache(projectName, signature string) error
	UnlockStageCache(projectName, signature string) error

	//ReleaseAllStageLocks() error
	//LockAllImagesReadOnly(projectName string) error
	//UnlockAllImages(projectName string) error
}
