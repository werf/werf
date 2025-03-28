package background

import (
	"os"

	chart "github.com/werf/common-go/pkg/lock"
	"github.com/werf/lockgate"
)

func IsBackgroundModeEnabled() bool {
	return os.Getenv("_WERF_BACKGROUND_MODE_ENABLED") == "1"
}

// TryLock tries background host-lock or panics if error.
func TryLock() bool {
	locker, err := chart.HostLocker()
	if err != nil {
		panic(err)
	}
	lockName := "werf.background-process"
	lockOptions := lockgate.AcquireOptions{NonBlocking: true}
	ok, _, err := locker.Acquire(lockName, lockOptions)
	if err != nil {
		panic(err)
	}
	return ok
}
