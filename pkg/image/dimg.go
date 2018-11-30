package image

type Dimg struct {
	*Stage
}

func NewDimgImage(fromImage *Stage, name string) *Dimg {
	return &Dimg{Stage: NewStageImage(fromImage, name)}
}

func (i *Dimg) Tag() error {
	return i.Stage.Tag(i.name)
}

func (i *Dimg) Export() error {
	return i.Stage.Export(i.name)
}
