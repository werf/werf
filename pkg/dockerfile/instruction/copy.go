package instruction

type Copy struct {
	*Base

	From  string
	Src   []string
	Dst   string
	Chown string
	Chmod string
}

func NewCopy(raw, from string, src []string, dst, chown, chmod string) *Copy {
	return &Copy{Base: NewBase(raw), From: from, Src: src, Dst: dst, Chown: chown, Chmod: chmod}
}

func (i *Copy) Name() string {
	return "COPY"
}
