package storage

import "github.com/werf/lockgate"

type LockManager interface {
	LockStage(projectName, signature string) (LockHandle, error)
	LockStageCache(projectName, signature string) (LockHandle, error)
	LockImage(projectName, imageName string) (LockHandle, error)
	LockStagesAndImages(projectName string, opts LockStagesAndImagesOptions) (LockHandle, error)
	Unlock(lockHandle LockHandle) error
}

type LockHandle struct {
	ProjectName    string              `json:"projectName"`
	LockgateHandle lockgate.LockHandle `json:"lockgateHandle"`
}

type LockStagesAndImagesOptions struct {
	GetOrCreateImagesOnly bool `json:"getOrCreateImagesOnly"`
}
