package instruction

type Cmd struct {
	Cmd []string
}

func NewCmd(cmd []string) *Cmd {
	return &Cmd{Cmd: cmd}
}

func (i *Cmd) Name() string {
	return "CMD"
}
