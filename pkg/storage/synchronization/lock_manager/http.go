package lock_manager

import (
	"context"
	"fmt"

	"github.com/werf/common-go/pkg/locker_with_retry"
	"github.com/werf/lockgate/pkg/distributed_locker"
)

func NewHttp(ctx context.Context, address, clientID string) (Interface, error) {
	url := fmt.Sprintf("%s/%s/locker", address, clientID)
	locker := distributed_locker.NewHttpLocker(url)
	lockerWithRetry := locker_with_retry.NewLockerWithRetry(ctx, locker, locker_with_retry.LockerWithRetryOptions{MaxAcquireAttempts: maxAcquireAttempts, MaxReleaseAttempts: maxReleaseAttempts})

	return NewGeneric(lockerWithRetry), nil
}
