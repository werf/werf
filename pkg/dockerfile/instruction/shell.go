package instruction

type Shell struct {
	*Base

	Shell []string
}

func NewShell(raw string, shell []string) *Shell {
	return &Shell{Base: NewBase(raw), Shell: shell}
}

func (i *Shell) Name() string {
	return "SHELL"
}
