package storage

import (
	"errors"
	"strings"
)

var (
	ErrBrokenImage   = errors.New("broken image")
	ErrImageNotFound = errors.New("image not found")
	ErrNotSupported  = errors.New("not supported")
	ErrStageNotFound = errors.New("stage not found")
	ErrStageRejected = errors.New("stage rejected")
)

func IsErrBrokenImage(err error) bool {
	return err != nil && strings.HasSuffix(err.Error(), ErrBrokenImage.Error())
}

func IsErrStageNotFound(err error) bool {
	return errors.Is(err, ErrStageNotFound)
}

func IsErrStageUnavailable(err error) bool {
	return errors.Is(err, ErrStageNotFound) || errors.Is(err, ErrBrokenImage) || errors.Is(err, ErrStageRejected)
}
