package stages

import (
	"context"

	"github.com/werf/werf/v2/pkg/deploy/bundles"
	"github.com/werf/werf/v2/pkg/docker_registry"
)

type CopyOptions struct {
	FromRegistryClient docker_registry.Interface
	ToRegistryClient   docker_registry.Interface
}

func Copy(ctx context.Context, fromAddr *bundles.Addr, toAddr *bundles.Addr, opts CopyOptions) error {
	fromStorage := NewStorageAccessor(fromAddr)
	toStorage := NewStorageAccessor()

	return fromStorage.CopyTo(ctx, toStorage, copyToOptions{})
}
