package instruction

type Workdir struct {
	*Base

	Workdir string
}

func NewWorkdir(raw, workdir string) *Workdir {
	return &Workdir{Base: NewBase(raw), Workdir: workdir}
}

func (i *Workdir) Name() string {
	return "WORKDIR"
}
