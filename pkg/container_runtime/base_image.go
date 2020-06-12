package container_runtime

import (
	"fmt"

	"github.com/werf/werf/pkg/image"

	"github.com/docker/docker/api/types"
	"github.com/werf/werf/pkg/docker"
)

type baseImage struct {
	name             string
	inspect          *types.ImageInspect
	stageDesc *image.StageDescription

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

func (i *baseImage) SetStageDescription(stageDesc *image.StageDescription) {
	i.stageDesc = stageDesc
}

func (i *baseImage) GetStageDescription() *image.StageDescription {
	return i.stageDesc
}

func (i *baseImage) IsExistsLocally() bool {
	return i.inspect != nil
}
