package storage

import (
	"context"

	"github.com/werf/lockgate"
)

type LockManager interface {
	LockStage(ctx context.Context, projectName, digest string) (LockHandle, error)
	Unlock(ctx context.Context, lockHandle LockHandle) error
}

type LockHandle struct {
	ProjectName    string              `json:"projectName"`
	LockgateHandle lockgate.LockHandle `json:"lockgateHandle"`
}

type LockStagesAndImagesOptions struct {
	GetOrCreateImagesOnly bool `json:"getOrCreateImagesOnly"`
}
