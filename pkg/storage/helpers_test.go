package storage

import (
	"context"
	"fmt"

	"github.com/google/go-containerregistry/pkg/v1/remote/transport"

	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/image"
)

var _ docker_registry.Interface = (*metadataPushRegistry)(nil)

type metadataPushRegistry struct {
	*pushImageRegistryStub
}

func (r *metadataPushRegistry) Tags(_ context.Context, _ string, _ ...docker_registry.Option) ([]string, error) {
	return nil, fmt.Errorf("tag listing must not be needed to publish metadata")
}

var _ docker_registry.Interface = (*stageLookupRegistry)(nil)

type stageLookupRegistry struct {
	*markerRegistry
	brokenImage *image.Info
}

func (r *stageLookupRegistry) GetRepoImage(ctx context.Context, reference string) (*image.Info, error) {
	info, err := r.markerRegistry.GetRepoImage(ctx, reference)
	if err != nil {
		return nil, err
	}
	if info == nil {
		return nil, fmt.Errorf("%s: %s", transport.ManifestUnknownErrorCode, reference)
	}
	if info == r.brokenImage {
		return nil, fmt.Errorf("%s: %s", transport.BlobUnknownErrorCode, reference)
	}
	return info, nil
}

var _ container_backend.ContainerBackend = (*stageLookupBackend)(nil)

type stageLookupBackend struct {
	container_backend.ContainerBackend
	info *image.Info
}

func (b *stageLookupBackend) GetImageInfo(_ context.Context, _ string, _ container_backend.GetImageInfoOpts) (*image.Info, error) {
	return b.info, nil
}
