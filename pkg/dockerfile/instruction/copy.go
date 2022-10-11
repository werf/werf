package instruction

type Copy struct {
	From string
	Src  []string
	Dst  string
}

func NewCopy(from string, src []string, dst string) *Copy {
	return &Copy{From: from, Src: src, Dst: dst}
}

func (i *Copy) Name() string {
	return "COPY"
}
