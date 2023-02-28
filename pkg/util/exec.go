package util

import (
	"os"
	"os/exec"
	"strings"
)

func ExecWerfBinaryCmd(args ...string) *exec.Cmd {
	cmd := exec.Command(strings.TrimSuffix(os.Args[0], "-in-a-user-namespace"), args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd
}

func ExecKubectlCmd(args ...string) *exec.Cmd {
	return ExecWerfBinaryCmd(append([]string{"kubectl"}, args...)...)
}
