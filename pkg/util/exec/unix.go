//go:build linux || darwin

package exec

import (
	"os/exec"
	"syscall"

	"github.com/werf/werf/v2/pkg/util/option"
)

// MakeDetachable ensures command could start a detachable process.
// Detachable process is a group leader process which has not controlling terminal.
func MakeDetachable(cmd *exec.Cmd) *exec.Cmd {
	attr := option.PtrValueOrDefault(cmd.SysProcAttr, syscall.SysProcAttr{})

	// Creates a new session if the calling process is not a process group leader and sets the process group ID.
	// Ensures the process has not controlling terminal.
	// Note. A process group leader is a process whose process group ID equals its PID.
	// https://man7.org/linux/man-pages/man2/setsid.2.html
	attr.Setsid = true

	cmd.SysProcAttr = &attr

	return cmd
}
