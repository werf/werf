package container_backend

import (
	"errors"
	"regexp"

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

var ErrPatchApply = errors.New(`cannot update source code added by git directive because the files being patched were modified by user commands in earlier stages (install, beforeSetup or setup)

Possible solutions:

  - If these files should not be changed, update the commands that modify them.

  - If these files should be changed, declare them explicitly using the stageDependencies.<install|beforeSetup|setup> directive. This ensures the files are updated before running user commands.

  - If these files are not needed, exclude them using the includePaths or excludePaths options under the git directive.`)

var dockerRateLimitCredsRe = regexp.MustCompile(
	`(?i)you have reached your pull rate limit as\s+'[^']+':.*?(?:\.|$)`,
)

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

func SanitizeError(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	sanitized := dockerRateLimitCredsRe.ReplaceAllString(
		msg,
		"You have reached your pull rate limit (credentials hidden).",
	)

	if sanitized == msg {
		return err
	}

	return errors.New(sanitized)
}
