package image

type Build struct {
	*Base
}

func NewBuildImage(id string) *Build {
	image := &Build{}
	image.Base = NewBaseImage(id)
	return image
}
