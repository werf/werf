package instruction

type Entrypoint struct {
	Entrypoint   []string
	PrependShell bool
}

func NewEntrypoint(entrypoint []string, prependShell bool) *Entrypoint {
	return &Entrypoint{Entrypoint: entrypoint, PrependShell: prependShell}
}

func (i *Entrypoint) Name() string {
	return "ENTRYPOINT"
}
