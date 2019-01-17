package build

import "os"

func debug() bool {
	return os.Getenv("WERF_BUILD_DEBUG") == "1"
}
