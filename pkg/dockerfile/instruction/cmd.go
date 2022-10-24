package instruction

type Cmd struct {
	*Base

	Cmd          []string
	PrependShell bool
}

func NewCmd(raw string, cmd []string, prependShell bool) *Cmd {
	return &Cmd{Base: NewBase(raw), Cmd: cmd, PrependShell: prependShell}
}

func (i *Cmd) Name() string {
	return "CMD"
}
