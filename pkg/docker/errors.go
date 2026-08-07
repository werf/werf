package docker

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

var contentNotFoundRe = regexp.MustCompile(`content digest (sha256:[a-f0-9]+): not found`)

func IsErrContentNotFound(err error) bool {
	if err == nil {
		return false
	}
	cause := errors.Cause(err)
	msg := cause.Error()
	// Match Docker/containerd "content digest ... not found" (containerd-snapshotter missing blob)
	// and generic Docker API "NotFound" (HTTP 404) responses.
	return (strings.Contains(msg, "content digest") || strings.Contains(msg, "NotFound")) &&
		strings.Contains(msg, "not found")
}

func ContentNotFoundDigest(err error) string {
	if err == nil {
		return ""
	}
	matches := contentNotFoundRe.FindStringSubmatch(errors.Cause(err).Error())
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

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

// IsErrPruneRunning returns true if err is "a prune operation is already running" error
// https://github.com/moby/moby/blob/25.0/volume/service/service.go#L212
// https://github.com/moby/moby/blob/25.0/daemon/images/image_prune.go#L38
func IsErrPruneRunning(err error) bool {
	if err == nil {
		return false
	}
	cause := errors.Cause(err)
	return strings.HasSuffix(cause.Error(), "a prune operation is already running")
}

func IsContainerNameConflict(err error) bool {
	if err == nil {
		return false
	}
	cause := errors.Cause(err)
	return strings.Contains(cause.Error(), "Conflict") &&
		strings.Contains(cause.Error(), "container name") &&
		strings.Contains(cause.Error(), "is already in use")
}
