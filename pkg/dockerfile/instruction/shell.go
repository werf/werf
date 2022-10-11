package instruction

type Shell struct {
	Shell []string
}

func NewShell(shell []string) *Shell {
	return &Shell{Shell: shell}
}

func (i *Shell) Name() string {
	return "SHELL"
}
