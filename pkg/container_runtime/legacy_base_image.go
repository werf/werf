package container_runtime

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"

	"github.com/werf/werf/pkg/image"
)

type legacyBaseImage struct {
	name      string
	inspect   *types.ImageInspect
	stageDesc *image.StageDescription

	DockerServerRuntime *DockerServerRuntime
}

func newLegacyBaseImage(name string, dockerServerRuntime *DockerServerRuntime) *legacyBaseImage {
	img := &legacyBaseImage{}
	img.name = name
	img.DockerServerRuntime = dockerServerRuntime
	return img
}

func (i *legacyBaseImage) Name() string {
	return i.name
}

func (i *legacyBaseImage) SetName(name string) {
	i.name = name
}

func (i *legacyBaseImage) MustResetInspect(ctx context.Context) error {
	if inspect, err := i.DockerServerRuntime.GetImageInspect(ctx, i.Name()); err != nil {
		return fmt.Errorf("unable to get inspect for image %s: %s", i.Name(), err)
	} else {
		i.SetInspect(inspect)
	}

	if i.inspect == nil {
		panic(fmt.Sprintf("runtime error: inspect must be (%s)", i.name))
	}
	return nil
}

func (i *legacyBaseImage) GetInspect() *types.ImageInspect {
	return i.inspect
}

func (i *legacyBaseImage) SetInspect(inspect *types.ImageInspect) {
	i.inspect = inspect
}

func (i *legacyBaseImage) UnsetInspect() {
	i.inspect = nil
}

func (i *legacyBaseImage) SetStageDescription(stageDesc *image.StageDescription) {
	i.stageDesc = stageDesc
}

func (i *legacyBaseImage) GetStageDescription() *image.StageDescription {
	return i.stageDesc
}

func (i *legacyBaseImage) IsExistsLocally() bool {
	return i.inspect != nil
}
