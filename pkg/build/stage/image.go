package stage

type Image interface {
	ReadDockerState() error
	IsImageExists() bool
	GetLabels() map[string]string
	AddServiceChangeLabel(name, value string)
}

type StubImage struct {
	Labels              map[string]string
	ServiceChangeLabels map[string]string
}

func (image *StubImage) GetLabels() map[string]string {
	return image.Labels
}

func (image *StubImage) AddServiceChangeLabel(name, value string) {
	image.ServiceChangeLabels[name] = value
}
