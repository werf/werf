package instruction

type Entrypoint struct {
	Entrypoint []string
}

func NewEntrypoint(entrypoint []string) *Entrypoint {
	return &Entrypoint{Entrypoint: entrypoint}
}

func (i *Entrypoint) Name() string {
	return "ENTRYPOINT"
}
