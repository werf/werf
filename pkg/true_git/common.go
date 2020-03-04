package true_git

import (
	"bytes"
	"io"
	"os/exec"
)

func setCommandRecordingLiveOutput(cmd *exec.Cmd) *bytes.Buffer {
	recorder := &bytes.Buffer{}

	if liveGitOutput {
		cmd.Stdout = io.MultiWriter(recorder, outStream)
		cmd.Stderr = io.MultiWriter(recorder, errStream)
	} else {
		cmd.Stdout = recorder
		cmd.Stderr = recorder
	}

	return recorder
}
