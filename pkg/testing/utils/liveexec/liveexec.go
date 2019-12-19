package liveexec

// ExecCommandOptions is an options for ExecCommand
type ExecCommandOptions struct {
	Env               map[string]string
	OutputLineHandler func(string)
}

// ExecCommand allows handling output of executed command in realtime by CommandOptions.OutputLineHandler.
// User could set expectations on the output lines in the CommandOptions.OutputLineHandler to fail fast
// and give immediate feedback of failed assertion during command execution.
func ExecCommand(dir, binPath string, opts ExecCommandOptions, arg ...string) error {
	return doExecCommand(dir, binPath, opts, arg...)
}
