package container_runtime

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/image"
)

type legacyBaseImage struct {
	name      string
	info      *image.Info
	stageDesc *image.StageDescription

	ContainerRuntime ContainerRuntime
}

func newLegacyBaseImage(name string, containerRuntime ContainerRuntime) *legacyBaseImage {
	img := &legacyBaseImage{}
	img.name = name
	img.ContainerRuntime = containerRuntime
	return img
}

func (i *legacyBaseImage) Name() string {
	return i.name
}

func (i *legacyBaseImage) SetName(name string) {
	i.name = name
}

func (i *legacyBaseImage) MustResetInfo(ctx context.Context) error {
	if info, err := i.ContainerRuntime.GetImageInfo(ctx, i.Name(), GetImageInfoOpts{}); err != nil {
		return fmt.Errorf("unable to get info for image %s: %s", i.Name(), err)
	} else {
		i.SetInfo(info)
	}

	if i.info == nil {
		panic(fmt.Sprintf("runtime error: info must be set for image %q", i.name))
	}
	return nil
}

func (i *legacyBaseImage) GetInfo() *image.Info {
	return i.info
}

func (i *legacyBaseImage) SetInfo(info *image.Info) {
	i.info = info
}

func (i *legacyBaseImage) UnsetInfo() {
	i.info = nil
}

func (i *legacyBaseImage) SetStageDescription(stageDesc *image.StageDescription) {
	i.stageDesc = stageDesc
}

func (i *legacyBaseImage) GetStageDescription() *image.StageDescription {
	return i.stageDesc
}

func (i *legacyBaseImage) IsExistsLocally() bool {
	return i.info != nil
}
