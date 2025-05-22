package container_backend

import (
	"errors"

	"github.com/containers/storage/types"
	"github.com/docker/cli/cli"
)

var (
	ErrUnsupportedFeature           = errors.New("unsupported feature")
	ErrCannotRemovePausedContainer  = errors.New("cannot remove paused container")
	ErrCannotRemoveRunningContainer = errors.New("cannot remove running container")
	ErrImageUsedByContainer         = types.ErrImageUsedByContainer
	ErrPruneIsAlreadyRunning        = errors.New("a prune operation is already running")
)

var ErrPatchApply = errors.New(`werf cannot apply the patch to the current source code because the files being added were modified by user commands in earlier stages.

- If these files should NOT be changed, update the instructions for the preceding stages with user commands.

- If these files SHOULD be changed, declare this dependency using the stageDependencies directive, and these files will be updated before running user commands.

- If these files are NOT required, exclude them using the git.[*].includePaths / excludePaths directives.`)

const (
	ErrPatchApplyCode = 42
)

var errByCode = map[int]error{
	ErrPatchApplyCode: ErrPatchApply,
}

func CliErrorByCode(err error) error {
	if err == nil {
		return nil
	}
	var statusError cli.StatusError
	if errors.As(err, &statusError) {
		if e, ok := errByCode[statusError.StatusCode]; ok {
			return e
		}
	}
	return err
}
