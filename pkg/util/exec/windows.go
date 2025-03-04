//go:build windows

package exec

import (
	"os/exec"
)

// MakeDetachable ensures command could start detachable process.
func MakeDetachable(cmd *exec.Cmd) *exec.Cmd {
	// We don't know right now there is or there is not the process detaching problem on Windows.
	// Moreover, syscall.SysProcAttr for windows has not "Setsid" (linux specific) field.
	// So we keep cross-platform compatibility do nothing.

	return cmd
}
