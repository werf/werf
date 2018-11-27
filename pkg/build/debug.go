package build

import "os"

func debug() bool {
	return os.Getenv("DAPP_BUILD_DEBUG") == "1"
}
