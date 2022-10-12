package instruction

type Add struct {
	Src   []string
	Dst   string
	Chown string
	Chmod string
}

func NewAdd(src []string, dst, chown, chmod string) *Add {
	return &Add{Src: src, Dst: dst, Chown: chown, Chmod: chmod}
}

func (i *Add) Name() string {
	return "ADD"
}
