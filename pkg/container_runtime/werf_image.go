package container_runtime

type WerfImage struct {
	*StageImage
}

func NewWerfImage(fromImage *StageImage, name string, localDockerServerRuntime *LocalDockerServerRuntime) *WerfImage {
	return &WerfImage{StageImage: NewStageImage(fromImage, name, localDockerServerRuntime)}
}

func (i *WerfImage) Tag() error {
	return i.StageImage.Tag(i.name)
}

func (i *WerfImage) Export() error {
	return i.StageImage.Export(i.name)
}
