package host_cleaning

import (
	"context"

	chart "github.com/werf/common-go/pkg/lock"
	"github.com/werf/lockgate"
)

func withHostLock(ctx context.Context, lockName string, fn func() error) error {
	lockOptions := lockgate.AcquireOptions{NonBlocking: true}
	return chart.WithHostLock(ctx, lockName, lockOptions, fn)
}
