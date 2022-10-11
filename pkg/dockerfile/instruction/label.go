package instruction

type Label struct {
	Labels map[string]string
}

func NewLabel(labels map[string]string) *Label {
	return &Label{Labels: labels}
}

func (i *Label) Name() string {
	return "LABEL"
}
