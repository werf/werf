package storage

import "github.com/flant/werf/pkg/image"

type StagesStorageCache interface {
	//type ImageInspect struct{}
	//GetImageInspect(imageName string) (*ImageInspect, error)
	//SetImageInspect(imageName string, inspect *ImageInspect) error

	GetImagesBySignature(projectName, signature string) (bool, []*image.Info, error)
	StoreImagesBySignature(projectName, signature string, imageInfo []*image.Info) error
}
