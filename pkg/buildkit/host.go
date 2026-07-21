package buildkit

import (
	"fmt"
	"os"
)

func HostFromEnv() string {
	if host := os.Getenv("WERF_BUILDKIT_HOST"); host != "" {
		return host
	}
	return os.Getenv("BUILDKIT_HOST")
}

func GetHost() (string, error) {
	host := HostFromEnv()
	if host == "" {
		return "", fmt.Errorf(`buildkit host is not specified: set $WERF_BUILDKIT_HOST or $BUILDKIT_HOST to a buildkitd endpoint (unix://, tcp://, docker-container://, kube-pod://, podman-container:// or ssh://), e.g. run "docker run -d --name buildkitd --privileged moby/buildkit" and set BUILDKIT_HOST=docker-container://buildkitd`)
	}
	return host, nil
}
