package instruction

type Add struct {
	Src []string
	Dst string
}

func NewAdd(src []string, dst string) *Add {
	return &Add{Src: src, Dst: dst}
}

func (i *Add) Name() string {
	return "ADD"
}
