package exec

import (
	"context"
	"errors"
	"os/exec"
	"syscall"
	"time"

	"github.com/werf/common-go/pkg/graceful"
)

// PrepareGracefulCancellation returns cmd which is ready for graceful cancellation.
//
// Graceful cancellation means that command has
//
//	a) Cancel() function to send signal termination signal. We use SIGTERM (instead of SIGKILL which is by default).
//	b) WaitDelay duration to limit time for cancellation. We use 5 minutes (by default time is not limited).
func PrepareGracefulCancellation(cmd *exec.Cmd) *exec.Cmd {
	cmd.Cancel = func() error {
		return cmd.Process.Signal(syscall.SIGTERM)
	}
	cmd.WaitDelay = time.Minute * 5
	return cmd
}

// CommandContextCancellation does exec.CommandContext() and PrepareGracefulCancellation.
func CommandContextCancellation(ctx context.Context, name string, arg ...string) *exec.Cmd {
	return PrepareGracefulCancellation(exec.CommandContext(ctx, name, arg...))
}

func TerminateIfCanceled(ctx context.Context, err error, exitCode int) {
	if errors.Is(ctx.Err(), context.Canceled) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
		graceful.Terminate(ctx, err, exitCode)
	}
}

// ExitCode derives exit code from cmd error.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}

	var exitErr *exec.ExitError

	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	} else {
		return 1
	}
}
