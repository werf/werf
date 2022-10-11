package instruction

type Expose struct {
	Ports []string
}

func NewExpose(ports []string) *Expose {
	return &Expose{Ports: ports}
}

func (i *Expose) Name() string {
	return "EXPOSE"
}
