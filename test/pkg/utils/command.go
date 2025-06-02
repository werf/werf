package utils

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/util/option"
	werfExec "github.com/werf/werf/v2/pkg/werf/exec"
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

	CancelOnOutput        string
	CancelOnOutputTimeout time.Duration
}

func RunCommandWithOptions(ctx context.Context, dir, command string, args []string, options RunCommandOptions) ([]byte, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd = werfExec.PrepareGracefulCancellation(cmd)

	if dir != "" {
		cmd.Dir = dir
	}

	cmd.Env = append(os.Environ(), options.ExtraEnv...)

	if options.ToStdin != "" {
		cmd.Stdin = bytes.NewReader([]byte(options.ToStdin))
	}

	stdout, err := cmd.StdoutPipe()
	Expect(err).To(Succeed())

	stderr, err := cmd.StderrPipe()
	Expect(err).To(Succeed())

	var outputReader io.Reader

	if options.NoStderr {
		outputReader = stdout
	} else {
		outputReader = io.MultiReader(stdout, stderr)
	}

	res := &bytes.Buffer{}

	Expect(cmd.Start()).To(Succeed())

	if options.CancelOnOutput != "" {
		copyReader := io.TeeReader(outputReader, res)
		waitForOutput(copyReader, options.CancelOnOutput, options.CancelOnOutputTimeout)
		Expect(cmd.Cancel()).To(Succeed())
	}

	_, err = io.Copy(res, outputReader)
	Expect(err).To(Succeed())

	err = cmd.Wait()

	_, _ = GinkgoWriter.Write(res.Bytes())

	if options.ShouldSucceed {
		errorDesc := fmt.Sprintf("%[2]s %[3]s (dir: %[1]s)", dir, command, strings.Join(args, " "))
		Expect(err).ShouldNot(HaveOccurred(), errorDesc)
	}

	return res.Bytes(), err
}

func ShelloutPack(command string) string {
	return fmt.Sprintf("eval $(echo %s | base64 -d)", base64.StdEncoding.EncodeToString([]byte(command)))
}

// waitForOutput waits for output and exits early or exits by timeout
func waitForOutput(reader io.Reader, output string, timeout time.Duration) {
	scanner := bufio.NewScanner(reader)
	tmr := time.NewTimer(option.ValueOrDefault(timeout, time.Minute))

	for scanner.Scan() {
		select {
		case <-tmr.C:
			return
		default:
			if strings.Contains(scanner.Text(), output) {
				return
			}
		}
	}
}
