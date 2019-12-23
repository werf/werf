// +build windows

package liveexec

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/onsi/ginkgo"
)

func doExecCommand(dir, binPath string, opts ExecCommandOptions, arg ...string) error {
	cmd := exec.Command(binPath, arg...)
	cmd.Dir = dir

	cmd.Env = os.Environ()
	for k, v := range opts.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	absDir, _ := filepath.Abs(dir)
	_, _ = fmt.Fprintf(ginkgo.GinkgoWriter, "\r\n[DEBUG] COMMAND in %s: %s %s\r\n\r\n", absDir, binPath, strings.Join(arg, " "))

	res, err := cmd.CombinedOutput()
	lineBuf := make([]byte, 0, 4096)
	for _, b := range res {
		if b == '\n' {
			line := string(lineBuf)
			lineBuf = lineBuf[:0]

			_, _ = fmt.Fprintf(ginkgo.GinkgoWriter, "[DEBUG] OUTPUT LINE: %s\r\n", line)

			if opts.OutputLineHandler != nil {
				opts.OutputLineHandler(line)
			}

			continue
		}

		lineBuf = append(lineBuf, b)
	}

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("exit code %d", exitError.ExitCode())
		}
		return err
	}
	return nil
}
