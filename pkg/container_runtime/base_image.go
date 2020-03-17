package container_runtime

import (
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/image"
)

type baseImage struct {
	name    string
	inspect *types.ImageInspect
	imgInfo *image.Info

	LocalDockerServerRuntime *LocalDockerServerRuntime
}

func newBaseImage(name string, localDockerServerRuntime *LocalDockerServerRuntime) *baseImage {
	image := &baseImage{}
	image.name = name
	image.LocalDockerServerRuntime = localDockerServerRuntime
	return image
}

func (i *baseImage) Name() string {
	return i.name
}

func (i *baseImage) SetName(name string) {
	i.name = name
}

func (i *baseImage) MustResetInspect() error {
	if inspect, err := i.LocalDockerServerRuntime.GetImageInspect(i.Name()); err != nil {
		return fmt.Errorf("unable to get inspect for image %s: %s", i.Name(), err)
	} else {
		i.SetInspect(inspect)
	}

	if i.inspect == nil {
		panic(fmt.Sprintf("runtime error: inspect must be (%s)", i.name))
	}
	return nil
}

func (i *baseImage) GetInspect() *types.ImageInspect {
	return i.inspect
}

func (i *baseImage) SetInspect(inspect *types.ImageInspect) {
	i.inspect = inspect
}

func (i *baseImage) UnsetInspect() {
	i.inspect = nil
}

func (i *baseImage) Untag() error {
	if err := docker.CliRmi(i.name, "--force"); err != nil {
		return err
	}

	i.UnsetInspect()

	return nil
}

func (i *baseImage) SetImageInfo(imgInfo *image.Info) {
	i.imgInfo = imgInfo
}

func (i *baseImage) GetImageInfo() *image.Info {
	return i.imgInfo
}

func (i *baseImage) IsExistsLocally() bool {
	return i.inspect != nil
}
