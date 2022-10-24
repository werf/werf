package instruction

type Add struct {
	*Base

	Src   []string
	Dst   string
	Chown string
	Chmod string
}

func NewAdd(raw string, src []string, dst, chown, chmod string) *Add {
	return &Add{Base: NewBase(raw), Src: src, Dst: dst, Chown: chown, Chmod: chmod}
}

func (i *Add) Name() string {
	return "ADD"
}
