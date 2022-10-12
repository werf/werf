package instruction

type Copy struct {
	From  string
	Src   []string
	Dst   string
	Chown string
	Chmod string
}

func NewCopy(from string, src []string, dst, chown, chmod string) *Copy {
	return &Copy{From: from, Src: src, Dst: dst, Chown: chown, Chmod: chmod}
}

func (i *Copy) Name() string {
	return "COPY"
}
