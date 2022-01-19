package true_git

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"

	"github.com/werf/logboek"
)

func NewGitCmd(ctx context.Context, opts *GitCmdOptions, cliArgs ...string) GitCmd {
	if len(cliArgs) == 0 {
		panic("cliArgs required, but not provided")
	}

	if opts == nil {
		opts = &GitCmdOptions{}
	}

	gitCmd := GitCmd{
		OutBuf:    &bytes.Buffer{},
		ErrBuf:    &bytes.Buffer{},
		OutErrBuf: &bytes.Buffer{},
	}

	gitCmd.Cmd = exec.Command("git", append(getCommonGitOptions(), cliArgs...)...)

	gitCmd.Dir = opts.RepoDir

	stdoutBuffs := []io.Writer{gitCmd.OutBuf, gitCmd.OutErrBuf}
	stderrBuffs := []io.Writer{gitCmd.ErrBuf, gitCmd.OutErrBuf}

	if liveGitOutput {
		stdoutBuffs = append(stdoutBuffs, logboek.Context(ctx).OutStream())
		stderrBuffs = append(stderrBuffs, logboek.Context(ctx).ErrStream())
	}

	gitCmd.Stdout = io.MultiWriter(stdoutBuffs...)
	gitCmd.Stderr = io.MultiWriter(stderrBuffs...)

	return gitCmd
}

type GitCmdOptions struct {
	RepoDir string
}

type GitCmd struct {
	*exec.Cmd

	// We always write to all of these buffs, unlike with exec.Cmd.Stdout(Stderr)
	OutBuf    *bytes.Buffer
	ErrBuf    *bytes.Buffer
	OutErrBuf *bytes.Buffer
}

func (c *GitCmd) Run(ctx context.Context) error {
	if debug() || liveGitOutput {
		logboek.Context(ctx).Debug().LogF("Running command %q\n", c)
	}

	switch err := c.Cmd.Run(); err.(type) {
	case *exec.ExitError:
		return fmt.Errorf("error running command %q: %s\nStdout:\n%s\nStderr:\n%s", c, err, c.OutBuf, c.ErrBuf)
	case error:
		return fmt.Errorf("error running command %q: %s", c, err)
	}

	return nil
}
