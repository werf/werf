package true_git

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	werfExec "github.com/werf/werf/v2/pkg/werf/exec"
)

func NewGitCmd(ctx context.Context, opts *GitCmdOptions, cliArgs ...string) GitCmd {
	if len(cliArgs) == 0 {
		panic("cliArgs required, but not provided")
	}

	if opts == nil {
		opts = &GitCmdOptions{}
	}

	gitCmd := GitCmd{
		OutBuf:    util.NewGoroutineSafeBuffer(),
		ErrBuf:    util.NewGoroutineSafeBuffer(),
		OutErrBuf: util.NewGoroutineSafeBuffer(),
	}

	gitCmd.Cmd = werfExec.CommandContextCancellation(ctx, "git", append(getCommonGitOptions(), cliArgs...)...)

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
	OutBuf    *util.GoroutineSafeBuffer
	ErrBuf    *util.GoroutineSafeBuffer
	OutErrBuf *util.GoroutineSafeBuffer
}

func (c *GitCmd) Run(ctx context.Context) error {
	if debug() || liveGitOutput {
		logboek.Context(ctx).Debug().LogF("Running command %q\n", c)
	}

	if err := c.Cmd.Run(); err != nil {
		werfExec.TerminateIfCanceled(ctx)

		var errExit *exec.ExitError
		if errors.As(err, &errExit) {
			return fmt.Errorf("error running command %q: %w\nStdout:\n%s\nStderr:\n%s", c, err, c.OutBuf, c.ErrBuf)
		}

		return fmt.Errorf("error running command %q: %w", c, err)
	}

	return nil
}
