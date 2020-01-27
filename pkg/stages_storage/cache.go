package stages_storage

type ImageInspect struct{}

type Cache interface {
	GetImageInspect(imageName string) (*ImageInspect, error)
	SetImageInspect(imageName string, inspect *ImageInspect) error
}
