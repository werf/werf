package background

import (
	"os"
)

func IsBackgroundModeEnabled() bool {
	return os.Getenv("_WERF_BACKGROUND_MODE_ENABLED") == "1"
}
