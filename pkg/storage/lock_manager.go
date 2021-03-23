package storage

import (
	"context"

	"github.com/werf/lockgate"
)

type LockManager interface {
	LockStage(ctx context.Context, projectName, signature string) (LockHandle, error)
	LockStageCache(ctx context.Context, projectName, signature string) (LockHandle, error)
	LockImage(ctx context.Context, projectName, imageName string) (LockHandle, error)
	Unlock(ctx context.Context, lockHandle LockHandle) error
}

type LockHandle struct {
	ProjectName    string              `json:"projectName"`
	LockgateHandle lockgate.LockHandle `json:"lockgateHandle"`
}
