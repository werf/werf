package image

import (
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	"github.com/flant/dapp/pkg/docker"
)

type Base struct {
	Name    string
	Inspect *types.ImageInspect
}

func NewBaseImage(name string) *Base {
	image := &Base{}
	image.Name = name
	return image
}

func (i *Base) MustGetId() (string, error) {
	if inspect, err := i.MustGetInspect(); err == nil {
		return inspect.ID, nil
	} else {
		return "", err
	}
}

func (i *Base) MustGetInspect() (*types.ImageInspect, error) {
	if inspect, err := i.GetInspect(); err == nil && inspect != nil {
		return inspect, nil
	} else if err != nil {
		return nil, err
	} else {
		panic(fmt.Sprintf("runtime error: inspect must be (%s)", i.Name))
	}
}

func (i *Base) GetInspect() (*types.ImageInspect, error) {
	if i.Inspect == nil {
		if err := i.resetInspect(); err != nil {
			if client.IsErrNotFound(err) {
				return nil, nil
			} else {
				return nil, err
			}
		}
	}
	return i.Inspect, nil
}

func (i *Base) resetInspect() error {
	inspect, err := docker.ImageInspect(i.Name)
	if err != nil {
		return err
	}

	i.Inspect = inspect
	return nil
}

func (i *Base) UnsetInspect() {
	i.Inspect = nil
}
