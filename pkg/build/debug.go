package build

import "os"

func debug() bool {
	return os.Getenv("WERF_CONVEYOR_DEBUG") == "1"
}
