package container_runtime

import "context"

type WerfImage struct {
	*StageImage
}

func NewWerfImage(fromImage *StageImage, name string, localDockerServerRuntime *LocalDockerServerRuntime) *WerfImage {
	return &WerfImage{StageImage: NewStageImage(fromImage, name, localDockerServerRuntime)}
}

func (i *WerfImage) Tag(ctx context.Context) error {
	return i.StageImage.Tag(ctx, i.name)
}

func (i *WerfImage) Export(ctx context.Context) error {
	return i.StageImage.Export(ctx, i.name)
}
