package true_git

import (
	"bytes"
	"context"
	"io"
	"os/exec"

	"github.com/werf/logboek"
)

func SetCommandRecordingLiveOutput(ctx context.Context, cmd *exec.Cmd) *bytes.Buffer {
	recorder := &bytes.Buffer{}

	if liveGitOutput {
		cmd.Stdout = io.MultiWriter(recorder, logboek.Context(ctx).OutStream())
		cmd.Stderr = io.MultiWriter(recorder, logboek.Context(ctx).ErrStream())
	} else {
		cmd.Stdout = recorder
		cmd.Stderr = recorder
	}

	return recorder
}

func getCommonGitOptions() []string {
	return []string{"-c", "core.autocrlf=false", "-c", "gc.auto=0"}
}
