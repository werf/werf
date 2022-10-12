package instruction

type Cmd struct {
	Cmd          []string
	PrependShell bool
}

func NewCmd(cmd []string, prependShell bool) *Cmd {
	return &Cmd{Cmd: cmd, PrependShell: prependShell}
}

func (i *Cmd) Name() string {
	return "CMD"
}
