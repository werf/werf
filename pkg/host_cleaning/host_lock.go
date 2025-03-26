package host_cleaning

import (
	"context"
	"errors"

	chart "github.com/werf/common-go/pkg/lock"
	"github.com/werf/lockgate"
	"github.com/werf/logboek"
)

// withHostLockOrNothing executes callback function if "soft" (NonBlocking=true) lock is acquired. Otherwise, does nothing.
func withHostLockOrNothing(ctx context.Context, lockName string, callback func() error) error {
	lockOptions := lockgate.AcquireOptions{NonBlocking: true}

	acquired, lock, err := chart.AcquireHostLock(ctx, lockName, lockOptions)
	if err != nil {
		return err
	}

	if !acquired {
		logboek.Context(ctx).Warn().LogF("Ignore locked %s\n", lockName)
		return nil
	}

	return errors.Join(callback(), chart.ReleaseHostLock(lock)) // join non-nil errors or return nil
}
