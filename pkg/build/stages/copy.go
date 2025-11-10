package stages

import (
	"context"

	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/ref"
	"github.com/werf/werf/v2/pkg/storage/manager"
)

type CopyOptions struct {
	InsecureRegistry      *bool
	SkipTlsVerifyRegistry *bool
	All                   bool
	ProjectName           string
}

func Copy(ctx context.Context, fromAddr *ref.Addr, toAddr *ref.Addr, fromStorageManager *manager.StorageManager, toDockerRegistry docker_registry.Interface, opts CopyOptions) error {
	fromStorage, err := NewStorageAccessor(ctx, fromAddr, fromStorageManager, toDockerRegistry, StorageAccessorOptions{
		RegistryOptions: RegistryStorageOptions{
			InsecureRegistry:      opts.InsecureRegistry,
			SkipTlsVerifyRegistry: opts.SkipTlsVerifyRegistry,
		},
		ArchiveOptions: ArchiveStorageOptions{},
	})
	if err != nil {
		return err
	}

	toStorage, err := NewStorageAccessor(ctx, toAddr, fromStorageManager, toDockerRegistry, StorageAccessorOptions{
		RegistryOptions: RegistryStorageOptions{
			InsecureRegistry:      opts.InsecureRegistry,
			SkipTlsVerifyRegistry: opts.SkipTlsVerifyRegistry,
		},
		ArchiveOptions: ArchiveStorageOptions{},
	})
	if err != nil {
		return err
	}

	return fromStorage.CopyTo(ctx, toStorage, copyToOptions{
		All:         opts.All,
		ProjectName: opts.ProjectName,
	})
}
