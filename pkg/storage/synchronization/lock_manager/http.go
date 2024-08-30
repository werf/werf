package lock_manager

import (
	"context"
	"fmt"

	"github.com/werf/lockgate/pkg/distributed_locker"
	"github.com/werf/nelm/pkg/locker_with_retry"
)

func NewHttp(ctx context.Context, address, clientID string) (Interface, error) {
	url := fmt.Sprintf("%s/%s/locker", address, clientID)
	locker := distributed_locker.NewHttpLocker(url)
	lockerWithRetry := locker_with_retry.NewLockerWithRetry(ctx, locker, locker_with_retry.LockerWithRetryOptions{MaxAcquireAttempts: maxAcquireAttempts, MaxReleaseAttempts: maxReleaseAttempts})

	return NewGeneric(lockerWithRetry), nil
}
