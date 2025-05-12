package utils

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func RunCommand(ctx context.Context, dir, command string, args ...string) ([]byte, error) {
	return RunCommandWithOptions(ctx, dir, command, args, RunCommandOptions{ShouldSucceed: false})
}

func RunSucceedCommand(ctx context.Context, dir, command string, args ...string) {
	_, _ = RunCommandWithOptions(ctx, dir, command, args, RunCommandOptions{ShouldSucceed: true})
}

func SucceedCommandOutputString(ctx context.Context, dir, command string, args ...string) string {
	res, _ := RunCommandWithOptions(ctx, dir, command, args, RunCommandOptions{ShouldSucceed: true})
	return string(res)
}

type RunCommandOptions struct {
	ExtraEnv      []string
	ToStdin       string
	ShouldSucceed bool
	NoStderr      bool
}

func RunCommandWithOptions(ctx context.Context, dir, command string, args []string, options RunCommandOptions) ([]byte, error) {
	cmd := exec.CommandContext(ctx, command, args...)

	if dir != "" {
		cmd.Dir = dir
	}

	cmd.Env = append(os.Environ(), options.ExtraEnv...)

	if options.ToStdin != "" {
		var b bytes.Buffer
		b.Write([]byte(options.ToStdin))
		cmd.Stdin = &b
	}

	var res []byte
	var err error
	if options.NoStderr {
		res, err = cmd.Output()
	} else {
		res, err = cmd.CombinedOutput()
	}

	_, _ = GinkgoWriter.Write(res)

	if options.ShouldSucceed {
		errorDesc := fmt.Sprintf("%[2]s %[3]s (dir: %[1]s)", dir, command, strings.Join(args, " "))
		Expect(err).ShouldNot(HaveOccurred(), errorDesc)
	}

	return res, err
}

func ShelloutPack(command string) string {
	return fmt.Sprintf("eval $(echo %s | base64 -d)", base64.StdEncoding.EncodeToString([]byte(command)))
}
