package lock_manager

import (
	"context"

	"github.com/werf/lockgate"
)

type Interface interface {
	LockStage(ctx context.Context, projectName, digest string) (LockHandle, error)
	Unlock(ctx context.Context, lockHandle LockHandle) error
}

type LockHandle struct {
	ProjectName    string              `json:"projectName"`
	LockgateHandle lockgate.LockHandle `json:"lockgateHandle"`
}
