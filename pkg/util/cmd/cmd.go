package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/werf/logboek"
)

func NewCmd(_ context.Context, command string, cliArgs ...string) Cmd {
	if len(cliArgs) == 0 {
		panic("cliArgs required, but not provided")
	}

	cmd := Cmd{
		OutBuf:    &bytes.Buffer{},
		ErrBuf:    &bytes.Buffer{},
		OutErrBuf: &bytes.Buffer{},
	}

	cmd.Cmd = exec.Command(command, cliArgs...)

	stdoutBuffs := []io.Writer{cmd.OutBuf, cmd.OutErrBuf}
	stderrBuffs := []io.Writer{cmd.ErrBuf, cmd.OutErrBuf}

	cmd.Stdout = io.MultiWriter(stdoutBuffs...)
	cmd.Stderr = io.MultiWriter(stderrBuffs...)

	return cmd
}

type Cmd struct {
	*exec.Cmd

	// We always write to all of these buffs, unlike with exec.Cmd.Stdout(Stderr)
	OutBuf    *bytes.Buffer
	ErrBuf    *bytes.Buffer
	OutErrBuf *bytes.Buffer
}

func (c *Cmd) Run(ctx context.Context) error {
	if os.Getenv("WERF_DEBUG_CMD") == "1" {
		logboek.Context(ctx).Default().LogF("Running command %q\n", c)
	}

	if err := c.Cmd.Run(); err != nil {
		var errExit *exec.ExitError
		if errors.As(err, &errExit) {
			return fmt.Errorf("error running command %q: %w\nStdout:\n%s\nStderr:\n%s", c, err, c.OutBuf, c.ErrBuf)
		}

		return fmt.Errorf("error running command %q: %w", c, err)
	}

	return nil
}
