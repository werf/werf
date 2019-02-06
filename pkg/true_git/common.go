package true_git

import (
	"bytes"
	"io"
	"os/exec"
)

func setCommandRecordingLiveOutput(cmd *exec.Cmd) *bytes.Buffer {
	recorder := &bytes.Buffer{}
	cmd.Stdout = io.MultiWriter(recorder, outStream)
	cmd.Stderr = io.MultiWriter(recorder, errStream)
	return recorder
}
