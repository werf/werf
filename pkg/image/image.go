package image

type Image struct {
	*StageImage
}

func NewImage(fromImage *StageImage, name string) *Image {
	return &Image{StageImage: NewStageImage(fromImage, name)}
}

func (i *Image) Tag() error {
	return i.StageImage.Tag(i.name)
}

func (i *Image) Export() error {
	return i.StageImage.Export(i.name)
}
