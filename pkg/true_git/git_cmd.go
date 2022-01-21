package true_git

import (
	"bytes"
	"context"
	"io"
	"os/exec"

	"github.com/werf/logboek"
)

type Cmd2 interface {
	Name() string
	Args() []string
	Env() []string
	WorkDir() string
	Stdin() io.Reader
	Stdout() io.Writer
	Stderr() io.Writer

	SetArgs(args []string)
	SetName(name string)
	SetEnv(env []string)
	SetWorkDir(dir string)
	SetStdin(stdin io.Reader)
	SetStdout(stdin io.Writer)
	SetStderr(stdin io.Writer)

	String() string
	Run() error
}

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

	gitCmd.Cmd = execCommand("git", append(getCommonGitOptions(), cliArgs...)...)

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

func NewCmd(ctx context.Context) Cmd {
	out := &cmd{
		OutBuf:    &bytes.Buffer{},
		ErrBuf:    &bytes.Buffer{},
		OutErrBuf: &bytes.Buffer{},
	}

	gitCmd.Cmd = execCommand("git", append(getCommonGitOptions(), cliArgs...)...)

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

//
// type ExecCmd struct {
// 	name        string
// 	args        []string
// 	env         []string
// 	workDir     string
// 	stdIn       *bytes.Buffer
// 	stdOut      *bytes.Buffer
// 	stdErr      *bytes.Buffer
// 	stdCombined *bytes.Buffer
// }
//
// func (c *ExecCmd) Name() string {
// 	return c.name
// }
//
// func (c *ExecCmd) Args() []string {
// 	return c.args
// }
//
// func (c *ExecCmd) Env() []string {
// 	return c.env
// }
//
// func (c *ExecCmd) WorkDir() string {
// 	return c.workDir
// }
//
// func (c *ExecCmd) Stdin() io.Reader {
// 	return c.stdIn
// }
//
// func (c *ExecCmd) Stdout() io.Writer {
// 	return c.stdOut
// }
//
// func (c *ExecCmd) Stderr() io.Writer {
// 	return c.stdErr
// }
//
// func (c *ExecCmd) SetArgs(args []string) {
// 	panic("implement me")
// }
//
// func (c *ExecCmd) SetName(name string) {
// 	panic("implement me")
// }
//
// func (c *ExecCmd) SetEnv(env []string) {
// 	panic("implement me")
// }
//
// func (c *ExecCmd) SetWorkDir(dir string) {
// 	panic("implement me")
// }
//
// func (c *ExecCmd) SetStdin(stdin io.Reader) {
// 	panic("implement me")
// }
//
// func (c *ExecCmd) SetStdout(stdin io.Writer) {
// 	panic("implement me")
// }
//
// func (c *ExecCmd) SetStderr(stdin io.Writer) {
// 	panic("implement me")
// }
//
// func (c *ExecCmd) String() string {
// 	panic("implement me")
// }
//
// func (c *ExecCmd) Run() error {
// 	panic("implement me")
// }

// func (c *ExecCmd) Run(ctx context.Context) error {
// 	if debug() || liveGitOutput {
// 		logboek.Context(ctx).Debug().LogF("Running command %q\n", c)
// 	}
//
// 	switch err := c.Cmd.Run(); err.(type) {
// 	case *exec.ExitError:
// 		// return fmt.Errorf("error running command %q: %s\nStdout:\n%s\nStderr:\n%s", c, err, c.OutBuf, c.ErrBuf)
// 		return err
// 	case error:
// 		return fmt.Errorf("error running command %q: %s", c, err)
// 	}
//
// 	return nil
// }

type Cmd interface {
	Name() string
	Args() []string
	Env() []string
	WorkDir() string
	Stdin() io.Reader
	Stdout() io.Writer
	Stderr() io.Writer

	SetArgs(args []string)
	SetName(name string)
	SetEnv(env []string)
	SetWorkDir(dir string)
	SetStdin(stdin io.Reader)
	SetStdout(stdin io.Writer)
	SetStderr(stdin io.Writer)

	String() string
	Run() error
}

type TestCmd struct {
	name        string
	args        []string
	env         []string
	workDir     string
	stdIn       *bytes.Buffer
	stdOut      *bytes.Buffer
	stdErr      *bytes.Buffer
	stdCombined *bytes.Buffer
}

func (c *TestCmd) String() string {
	panic("not implemented")
}

func (c *TestCmd) Run() error {
	panic("not implemented")
}

type ExecCmd struct {
	cmd exec.Cmd
}

func (c *baseCmd) Name() string {
	return c.name
}

func (c *baseCmd) Args() []string {
	return c.args
}

func (c *baseCmd) Env() []string {
	return c.env
}

func (c *baseCmd) WorkDir() string {
	return c.workDir
}

func (c *baseCmd) Stdin() io.Reader {
	return c.stdIn
}

func (c *baseCmd) Stdout() io.Writer {
	return c.stdOut
}

func (c *baseCmd) Stderr() io.Writer {
	return c.stdErr
}
