package helm

import (
	"bytes"
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
