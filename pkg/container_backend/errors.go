package container_backend

import (
	"errors"

	"github.com/containers/storage/types"
)

var (
	ErrUnsupportedFeature           = errors.New("unsupported feature")
	ErrCannotRemovePausedContainer  = errors.New("cannot remove paused container")
	ErrCannotRemoveRunningContainer = errors.New("cannot remove running container")
	ErrImageUsedByContainer         = types.ErrImageUsedByContainer
)
