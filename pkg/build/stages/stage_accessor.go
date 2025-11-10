package stages

import (
	"context"
	"fmt"

	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/ref"
	"github.com/werf/werf/v2/pkg/storage/manager"
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
	InsecureRegistry      *bool
	SkipTlsVerifyRegistry *bool
}

type ArchiveStorageOptions struct {
}

func NewStorageAccessor(ctx context.Context, addr *ref.Addr, storageManager *manager.StorageManager, dockerRegistry docker_registry.Interface, opts StorageAccessorOptions) (StorageAccessor, error) {
	switch {
	case addr.RegistryAddress != nil:
		return NewRemoteStorage(addr.RegistryAddress, storageManager, dockerRegistry), nil
	case addr.ArchiveAddress != nil:
		return NewArchiveStorage(NewArchiveStorageFileReader(addr.ArchiveAddress.Path), NewArchiveStorageFileWriter(addr.ArchiveAddress.Path)), nil
	default:
		panic(fmt.Sprintf("invalid address given %#v", addr))
	}
}
