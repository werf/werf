package instruction

type Label struct {
	*Base

	Labels map[string]string
}

func NewLabel(raw string, labels map[string]string) *Label {
	return &Label{Base: NewBase(raw), Labels: labels}
}

func (i *Label) Name() string {
	return "LABEL"
}
