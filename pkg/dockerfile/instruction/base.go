package instruction

type Base struct {
	Raw string
}

func NewBase(raw string) *Base {
	return &Base{Raw: raw}
}
