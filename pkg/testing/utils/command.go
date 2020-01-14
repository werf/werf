package utils

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func RunCommand(dir, command string, args ...string) ([]byte, error) {
	return RunCommandWithOptions(dir, command, args, RunCommandOptions{ShouldSucceed: false})
}

func RunSucceedCommand(dir, command string, args ...string) {
	_, _ = RunCommandWithOptions(dir, command, args, RunCommandOptions{ShouldSucceed: true})
}

func SucceedCommandOutputString(dir, command string, args ...string) string {
	res, _ := RunCommandWithOptions(dir, command, args, RunCommandOptions{ShouldSucceed: true})
	return string(res)
}

type RunCommandOptions struct {
	ToStdin       string
	ShouldSucceed bool
}

func RunCommandWithOptions(dir, command string, args []string, options RunCommandOptions) ([]byte, error) {
	if command == werfBinPath {
		args = WerfBinArgs(args...)
	}

	cmd := exec.Command(command, args...)

	if dir != "" {
		cmd.Dir = dir
	}

	if options.ToStdin != "" {
		var b bytes.Buffer
		b.Write([]byte(options.ToStdin))
		cmd.Stdin = &b
	}

	res, err := cmd.CombinedOutput()
	_, _ = GinkgoWriter.Write(res)

	if options.ShouldSucceed {
		errorDesc := fmt.Sprintf("%[2]s %[3]s (dir: %[1]s)", dir, command, strings.Join(args, " "))
		Ω(err).ShouldNot(HaveOccurred(), errorDesc)
	}

	return res, err
}

func ShelloutPack(command string) string {
	return fmt.Sprintf("eval $(echo %s | base64 -d)", base64.StdEncoding.EncodeToString([]byte(command)))
}
