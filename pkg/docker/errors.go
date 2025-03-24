package docker

import (
	"strings"

	"github.com/pkg/errors"
)

// IsErrContainerPaused returns true if err is "container is paused" error
// https://github.com/moby/moby/blob/25.0/daemon/delete.go#L94
func IsErrContainerPaused(err error) bool {
	if err == nil {
		return false
	}
	cause := errors.Cause(err)
	if !strings.HasPrefix(cause.Error(), "cannot remove container") {
		return false
	}
	return strings.HasSuffix(cause.Error(), "container is paused and must be unpaused first")
}

// IsErrContainerRunning returns true if err is "container is running" error
// https://github.com/moby/moby/blob/25.0/daemon/delete.go#L96
func IsErrContainerRunning(err error) bool {
	if err == nil {
		return false
	}
	cause := errors.Cause(err)
	if !strings.HasPrefix(cause.Error(), "cannot remove container") {
		return false
	}
	return strings.HasSuffix(cause.Error(), "container is running: stop the container before removing or force remove")
}
