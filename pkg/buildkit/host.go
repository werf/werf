package buildkit

import (
	"os"
)

func HostFromEnv() string {
	if host := os.Getenv("WERF_BUILDKIT_HOST"); host != "" {
		return host
	}
	return os.Getenv("BUILDKIT_HOST")
}
