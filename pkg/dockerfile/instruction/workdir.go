package instruction

type Workdir struct {
	Workdir string
}

func NewWorkdir(workdir string) *Workdir {
	return &Workdir{Workdir: workdir}
}

func (i *Workdir) Name() string {
	return "WORKDIR"
}
