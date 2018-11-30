package image

type build struct {
	*base
}

func newBuildImage(id string) *build {
	image := &build{}
	image.base = newBaseImage(id)
	return image
}
