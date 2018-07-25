package build

type Image interface {
	GetLabels() map[string]string
}

type StubImage struct {
	Labels map[string]string
}

func (image *StubImage) GetLabels() map[string]string {
	return image.Labels
}
