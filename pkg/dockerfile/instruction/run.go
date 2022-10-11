package instruction

type Run struct {
	Command []string
}

func NewRun(command []string) *Run {
	return &Run{Command: command}
}

func (i *Run) Name() string {
	return "RUN"
}
