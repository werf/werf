package instruction

type Maintainer struct {
	Maintainer string
}

func (i *Maintainer) Name() string {
	return "MAINTAINER"
}
