package stages

import (
	"context"

	"github.com/werf/werf/v2/pkg/build"
	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/ref"
	"github.com/werf/werf/v2/pkg/storage/manager"
)

type CopyOptions struct {
	RegistryClient    docker_registry.Interface
	StorageManager    *manager.StorageManager
	ConveyorWithRetry *build.ConveyorWithRetryWrapper
	All               bool
	ProjectName       string
	BuildOptions      build.BuildOptions
}

func Copy(ctx context.Context, fromAddr, toAddr *ref.Addr, opts CopyOptions) error {
	from := NewStorageAccessor(fromAddr, StorageAccessorOptions{
		DockerRegistry:           opts.RegistryClient,
		StorageManager:           opts.StorageManager,
		ConveyorWithRetryWrapper: opts.ConveyorWithRetry,
	})

	to := NewStorageAccessor(toAddr, StorageAccessorOptions{})

	return from.CopyTo(ctx, to, copyToOptions{
		All:          opts.All,
		ProjectName:  opts.ProjectName,
		BuildOptions: opts.BuildOptions,
	})
}
