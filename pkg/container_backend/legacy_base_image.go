package container_backend

import (
	"context"
	"fmt"
	"sync"

	"github.com/werf/werf/v2/pkg/image"
)

type legacyBaseImage struct {
	name           string
	info           *image.Info
	stageDesc      *image.StageDesc
	finalStageDesc *image.StageDesc

	mu sync.RWMutex

	ContainerBackend ContainerBackend
}

func newLegacyBaseImage(name string, containerBackend ContainerBackend) *legacyBaseImage {
	img := &legacyBaseImage{}
	img.name = name
	img.ContainerBackend = containerBackend
	return img
}

func (i *legacyBaseImage) Name() string {
	return i.name
}

func (i *legacyBaseImage) SetName(name string) {
	i.name = name
}

func (i *legacyBaseImage) MustResetInfo(ctx context.Context) error {
	if info, err := i.ContainerBackend.GetImageInfo(ctx, i.Name(), GetImageInfoOpts{}); err != nil {
		return fmt.Errorf("unable to get info for image %s: %w", i.Name(), err)
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

func (i *legacyBaseImage) SetStageDesc(stageDesc *image.StageDesc) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.stageDesc = stageDesc
	i.SetInfo(stageDesc.Info)
}

func (i *legacyBaseImage) GetStageDesc() *image.StageDesc {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.stageDesc
}

func (i *legacyBaseImage) SetFinalStageDesc(stageDesc *image.StageDesc) {
	i.finalStageDesc = stageDesc
}

func (i *legacyBaseImage) GetFinalStageDesc() *image.StageDesc {
	return i.finalStageDesc
}

func (i *legacyBaseImage) IsExistsLocally() bool {
	return i.info != nil
}
