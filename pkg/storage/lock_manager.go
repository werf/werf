package storage

type LockManager interface {
	LockStage(projectName, signature string) error
	UnlockStage(projectName, signature string) error

	LockStageCache(projectName, signature string) error
	UnlockStageCache(projectName, signature string) error

	LockImage(imageName string) error
	UnlockImage(imageName string) error

	LockStagesAndImages(projectName string, opts LockStagesAndImagesOptions) error
	UnlockStagesAndImages(projectName string) error
}

type LockStagesAndImagesOptions struct {
	GetOrCreateImagesOnly bool
}
