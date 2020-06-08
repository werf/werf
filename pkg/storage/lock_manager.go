package storage

import "github.com/werf/lockgate"

type LockManager interface {
	LockStage(projectName, signature string) (LockHandle, error)
	LockStageCache(projectName, signature string) (LockHandle, error)
	LockImage(projectName, imageName string) (LockHandle, error)
	LockStagesAndImages(projectName string, opts LockStagesAndImagesOptions) (LockHandle, error)
	LockDeployProcess(projectName string, releaseName string, kubeContextName string) (LockHandle, error)
	Unlock(lock LockHandle) error
}

type LockHandle struct {
	ProjectName    string
	LockgateHandle lockgate.LockHandle
}
