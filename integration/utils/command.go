package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	. "github.com/onsi/gomega"
)

func RunCommand(dir, command string, args ...string) {
	errorDesc := fmt.Sprintf("%[2]s %[3]s (dir: %[1]s)", dir, command, strings.Join(args, " "))
	Ω(runCommand(dir, command, args...)).Should(Succeed(), errorDesc)
}

func runCommand(dir, bin string, args ...string) error {
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
