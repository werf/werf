package image

import (
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	"github.com/flant/werf/pkg/docker"
)

type base struct {
	name    string
	inspect *types.ImageInspect
}

func newBaseImage(name string) *base {
	image := &base{}
	image.name = name
	return image
}

func (i *base) Name() string {
	return i.name
}

func (i *base) MustGetId() (string, error) {
	if inspect, err := i.MustGetInspect(); err == nil {
		return inspect.ID, nil
	} else {
		return "", err
	}
}

func (i *base) MustGetInspect() (*types.ImageInspect, error) {
	if inspect, err := i.GetInspect(); err == nil && inspect != nil {
		return inspect, nil
	} else if err != nil {
		return nil, err
	} else {
		panic(fmt.Sprintf("runtime error: inspect must be (%s)", i.name))
	}
}

func (i *base) ResetInspect() error {
	i.unsetInspect()
	_, err := i.GetInspect()
	return err
}

func (i *base) GetInspect() (*types.ImageInspect, error) {
	if i.inspect == nil {
		if err := i.resetInspect(); err != nil {
			if client.IsErrNotFound(err) {
				return nil, nil
			} else {
				return nil, err
			}
		}
	}
	return i.inspect, nil
}

func (i *base) resetInspect() error {
	inspect, err := docker.ImageInspect(i.name)
	if err != nil {
		return err
	}

	i.inspect = inspect
	return nil
}

func (i *base) unsetInspect() {
	i.inspect = nil
}

func (i *base) Untag() error {
	if err := docker.CliRmi(i.name); err != nil {
		return err
	}

	i.unsetInspect()

	return nil
}
