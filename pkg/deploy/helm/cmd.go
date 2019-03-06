package helm

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strings"
)

var (
	KubeContext string
)

func HelmCmd(args ...string) (stdout string, stderr string, err error) {
	if KubeContext != "" {
		args = append([]string{"--kube-context", KubeContext}, args...)
	}

	cmd := exec.Command("helm", args...)
	cmd.Env = os.Environ()

	var stdoutBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	err = cmd.Run()
	stdout = strings.TrimSpace(stdoutBuf.String())
	stderr = strings.TrimSpace(stderrBuf.String())

	return
}

func FormatHelmCmdError(stdout, stderr string, err error) error {
	errMsg := err.Error()

	formattedHelmCmdOutput := FormatHelmCmdOutput(stdout, stderr)
	if formattedHelmCmdOutput != "" {
		errMsg += "\n"
		errMsg += formattedHelmCmdOutput
	}

	return errors.New(errMsg)
}

func FormatHelmCmdOutput(stdout, stderr string) string {
	var args []string
	for _, arg := range []string{stdout, stderr} {
		if arg != "" {
			args = append(args, arg)
		}
	}

	return strings.Join(args, "\n")
}
