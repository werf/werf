package stage

type Image interface {
	GetName() string
	ReadDockerState() error
	IsImageExists() bool
	GetLabels() map[string]string

	AddRunCommands([]string)
	AddEnv(map[string]interface{})
	AddServiceChangeLabel(name, value string)
	AddVolume(volume string)
}

type StubImage struct {
	Labels              map[string]string
	ServiceChangeLabels map[string]string
}

func (image *StubImage) GetLabels() map[string]string {
	return image.Labels
}

func (image *StubImage) AddRunCommands([]string) {}

func (image *StubImage) AddEnv(map[string]interface{}) {}

func (image *StubImage) AddVolume(string) {}

func (image *StubImage) AddServiceChangeLabel(name, value string) {
	image.ServiceChangeLabels[name] = value
}

func (image *StubImage) ReadDockerState() error {
	return nil
}

func (image *StubImage) IsImageExists() bool {
	return false
}

func (image *StubImage) GetName() string {
	return ""
}
