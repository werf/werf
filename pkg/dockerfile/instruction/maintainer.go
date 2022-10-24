package instruction

type Maintainer struct {
	*Base

	Maintainer string
}

func NewMaintainer(raw, maintainer string) *Maintainer {
	return &Maintainer{Base: NewBase(raw), Maintainer: maintainer}
}

func (i *Maintainer) Name() string {
	return "MAINTAINER"
}
