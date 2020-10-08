package storage

import (
	"context"

	"github.com/werf/lockgate"
)

type LockManager interface {
	LockStage(ctx context.Context, projectName, digest string) (LockHandle, error)
	LockStageCache(ctx context.Context, projectName, digest string) (LockHandle, error)
	LockImage(ctx context.Context, projectName, imageName string) (LockHandle, error)
	LockStagesAndImages(ctx context.Context, projectName string, opts LockStagesAndImagesOptions) (LockHandle, error)
	Unlock(ctx context.Context, lockHandle LockHandle) error
}

type LockHandle struct {
	ProjectName    string              `json:"projectName"`
	LockgateHandle lockgate.LockHandle `json:"lockgateHandle"`
}

type LockStagesAndImagesOptions struct {
	GetOrCreateImagesOnly bool `json:"getOrCreateImagesOnly"`
}
