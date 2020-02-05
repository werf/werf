package stages_storage

type StagesStorageCache interface {
	//type ImageInspect struct{}
	//GetImageInspect(imageName string) (*ImageInspect, error)
	//SetImageInspect(imageName string, inspect *ImageInspect) error

	GetImagesBySignature(projectName, signature string) (bool, []*ImageInfo, error)
	StoreImagesBySignature(projectName, signature string, imageInfo []*ImageInfo) error
}
