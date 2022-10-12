package instruction

import "github.com/opencontainers/runtime-spec/specs-go"

type Run struct {
	Command      []string
	Args         []string      // runtime args like --security and --network
	Mounts       []specs.Mount // structured --mount args
	PrependShell bool
}

func NewRun(command, args []string, mounts []specs.Mount, prependShell bool) *Run {
	return &Run{Command: command, Args: args, Mounts: mounts, PrependShell: prependShell}
}

func (i *Run) Name() string {
	return "RUN"
}
