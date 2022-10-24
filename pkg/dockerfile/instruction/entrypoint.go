package instruction

type Entrypoint struct {
	*Base

	Entrypoint   []string
	PrependShell bool
}

func NewEntrypoint(raw string, entrypoint []string, prependShell bool) *Entrypoint {
	return &Entrypoint{Base: NewBase(raw), Entrypoint: entrypoint, PrependShell: prependShell}
}

func (i *Entrypoint) Name() string {
	return "ENTRYPOINT"
}
