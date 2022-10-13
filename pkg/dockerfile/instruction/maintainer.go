package instruction

type Maintainer struct {
	Maintainer string
}

func NewMaintainer(maintainer string) *Maintainer {
	return &Maintainer{Maintainer: maintainer}
}

func (i *Maintainer) Name() string {
	return "MAINTAINER"
}
