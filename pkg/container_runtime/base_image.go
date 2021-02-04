package container_runtime

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"

	"github.com/werf/werf/pkg/image"
)

type baseImage struct {
	name      string
	inspect   *types.ImageInspect
	stageDesc *image.StageDescription

	LocalDockerServerRuntime *LocalDockerServerRuntime
}

func newBaseImage(name string, localDockerServerRuntime *LocalDockerServerRuntime) *baseImage {
	img := &baseImage{}
	img.name = name
	img.LocalDockerServerRuntime = localDockerServerRuntime
	return img
}

func (i *baseImage) Name() string {
	return i.name
}

func (i *baseImage) SetName(name string) {
	i.name = name
}

func (i *baseImage) MustResetInspect(ctx context.Context) error {
	if inspect, err := i.LocalDockerServerRuntime.GetImageInspect(ctx, i.Name()); err != nil {
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

func (i *baseImage) SetStageDescription(stageDesc *image.StageDescription) {
	i.stageDesc = stageDesc
}

func (i *baseImage) GetStageDescription() *image.StageDescription {
	return i.stageDesc
}

func (i *baseImage) IsExistsLocally() bool {
	return i.inspect != nil
}
