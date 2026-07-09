package container_backend

import (
	"errors"

	"github.com/werf/werf/v2/pkg/log_sanitize"
)

var (
	ErrUnsupportedFeature           = errors.New("unsupported feature")
	ErrCannotRemovePausedContainer  = errors.New("cannot remove paused container")
	ErrCannotRemoveRunningContainer = errors.New("cannot remove running container")
	ErrImageUsedByContainer         = errors.New("image is in use by a container")
	ErrPruneIsAlreadyRunning        = errors.New("a prune operation is already running")
)

func SanitizeError(err error) error {
	if err == nil {
		return nil
	}

	msg := err.Error()
	sanitized := log_sanitize.SanitizeDockerRateLimit(msg)

	if sanitized == msg {
		return err
	}

	return errors.New(sanitized)
}
