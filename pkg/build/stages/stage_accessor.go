package stages

import (
	"context"
	"fmt"

	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/ref"
)

type copyToOptions struct {
	All         bool
	ProjectName string
}

type StorageAccessor interface {
	CopyTo(ctx context.Context, to StorageAccessor, opts copyToOptions) error
	CopyFromArchive(ctx context.Context, fromArchive *ArchiveStorage, opts copyToOptions) error
	CopyFromRemote(ctx context.Context, fromRemote *RemoteStorage, opts copyToOptions) error
}

type StorageAccessorOptions struct {
	RegistryOptions RegistryStorageOptions
	ArchiveOptions  ArchiveStorageOptions
}

type RegistryStorageOptions struct {
	ProjectName                  string
	ContainerBackend             container_backend.ContainerBackend
	CommonCmdData                *common.CmdData
	DisableCleanup               bool
	DisableGitHistoryBasedPolicy bool
}

type ArchiveStorageOptions struct {
}

func NewStorageSrcAccessor(ctx context.Context, addr *ref.Addr, opts StorageAccessorOptions) (StorageAccessor, error) {
	switch {
	case addr.RegistryAddress != nil:
		opts.RegistryOptions.CommonCmdData.Repo.Address = &addr.RegistryAddress.Repo
		storageManager, err := common.NewStorageManager(ctx, &common.NewStorageManagerConfig{
			ProjectName:                    opts.RegistryOptions.ProjectName,
			ContainerBackend:               opts.RegistryOptions.ContainerBackend,
			CmdData:                        opts.RegistryOptions.CommonCmdData,
			CleanupDisabled:                opts.RegistryOptions.DisableCleanup,
			GitHistoryBasedCleanupDisabled: opts.RegistryOptions.DisableGitHistoryBasedPolicy,
		})
		if err != nil {
			return nil, err
		}

		dockerRegistry, err := common.CreateDockerRegistry(ctx, addr.Repo, *opts.RegistryOptions.CommonCmdData.InsecureRegistry, *opts.RegistryOptions.CommonCmdData.SkipTlsVerifyRegistry)
		if err != nil {
			return nil, err
		}

		return NewRemoteStorage(addr.RegistryAddress, storageManager, dockerRegistry), nil
	case addr.ArchiveAddress != nil:
		return NewArchiveStorage(NewArchiveStorageFileReader(addr.ArchiveAddress.Path), NewArchiveStorageFileWriter(addr.ArchiveAddress.Path)), nil
	default:
		panic(fmt.Sprintf("invalid address given %#v", addr))
	}
}

func NewStorageDstAccessor(ctx context.Context, addr *ref.Addr, opts StorageAccessorOptions) (StorageAccessor, error) {
	switch {
	case addr.RegistryAddress != nil:
		dockerRegistry, err := common.CreateDockerRegistry(ctx, addr.Repo, *opts.RegistryOptions.CommonCmdData.InsecureRegistry, *opts.RegistryOptions.CommonCmdData.SkipTlsVerifyRegistry)
		if err != nil {
			return nil, err
		}

		return NewRemoteStorage(addr.RegistryAddress, nil, dockerRegistry), nil
	case addr.ArchiveAddress != nil:
		return NewArchiveStorage(NewArchiveStorageFileReader(addr.ArchiveAddress.Path), NewArchiveStorageFileWriter(addr.ArchiveAddress.Path)), nil
	default:
		panic(fmt.Sprintf("invalid address given %#v", addr))
	}
}
