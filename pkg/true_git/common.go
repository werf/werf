package true_git

import (
	"bytes"
	"io"
	"os"
	"os/exec"
)

func setCommandRecordingLiveOutput(cmd *exec.Cmd) *bytes.Buffer {
	recorder := &bytes.Buffer{}
	cmd.Stdout = io.MultiWriter(recorder, os.Stdout)
	cmd.Stderr = io.MultiWriter(recorder, os.Stderr)
	return recorder
}
