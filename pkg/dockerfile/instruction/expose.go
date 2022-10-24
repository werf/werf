package instruction

type Expose struct {
	*Base

	Ports []string
}

func NewExpose(raw string, ports []string) *Expose {
	return &Expose{Base: NewBase(raw), Ports: ports}
}

func (i *Expose) Name() string {
	return "EXPOSE"
}
