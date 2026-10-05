package manager

import (
	"context"

	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/image"
)

type lookupBackend struct {
	container_backend.ContainerBackend
	images image.ImagesList
	info   *image.Info
}

var _ container_backend.ContainerBackend = (*lookupBackend)(nil)

func (backend *lookupBackend) Images(_ context.Context, _ container_backend.ImagesOptions) (image.ImagesList, error) {
	return backend.images, nil
}

func (backend *lookupBackend) GetImageInfo(_ context.Context, _ string, _ container_backend.GetImageInfoOpts) (*image.Info, error) {
	return backend.info, nil
}
