package stage

import "github.com/flant/dapp/pkg/image"

type StubImage struct {
	labels map[string]string
}

func (i *StubImage) Name() string {
	return ""
}

func (i *StubImage) Labels() map[string]string {
	return i.labels
}

func (i *StubImage) Container() image.Container {
	return nil
}

func (i *StubImage) BuilderContainer() image.BuilderContainer {
	return nil
}

func (i *StubImage) IsExists() (bool, error) {
	return false, nil
}

func (i *StubImage) ReadDockerState() error {
	return nil
}

func (i *StubImage) Pull() error {
	return nil
}
