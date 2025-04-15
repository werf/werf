package background

import (
	"os"

	"github.com/werf/lockgate"
	"github.com/werf/werf/v2/pkg/werf"
)

func IsBackgroundModeEnabled() bool {
	return os.Getenv("_WERF_BACKGROUND_MODE_ENABLED") == "1"
}

// TryLock tries background host-lock or panics if error.
//
// The logger is not available inside of this function
// because we use the function before initialization of the logger.
func TryLock() bool {
	// In background mode we always initialize werf using empty defaults.
	err := werf.Init("", "")
	if err != nil {
		panic(err)
	}
	lockName := "werf.background-process"
	lockOptions := lockgate.AcquireOptions{NonBlocking: true}
	ok, _, err := werf.HostLocker().Locker().Acquire(lockName, lockOptions)
	if err != nil {
		panic(err)
	}
	return ok
}
