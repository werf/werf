package stages

import (
	"context"
	"fmt"

	"github.com/werf/werf/v2/pkg/build"
	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/ref"
	"github.com/werf/werf/v2/pkg/storage/manager"
)

type copyToOptions struct {
	All          bool
	ProjectName  string
	BuildOptions build.BuildOptions
}

type StorageAccessor interface {
	CopyTo(ctx context.Context, to StorageAccessor, opts copyToOptions) error
	CopyFromArchive(ctx context.Context, fromArchive *ArchiveStorage, opts copyToOptions) error
	CopyFromRemote(ctx context.Context, fromRemote *RemoteStorage, opts copyToOptions) error
}

type StorageAccessorOptions struct {
	DockerRegistry           docker_registry.Interface
	StorageManager           *manager.StorageManager
	ConveyorWithRetryWrapper *build.ConveyorWithRetryWrapper
}

func NewStorageAccessor(addr *ref.Addr, opts StorageAccessorOptions) StorageAccessor {
	switch {
	case addr.RegistryAddress != nil:
		return NewRemoteStorage(addr.RegistryAddress, opts.DockerRegistry, opts.StorageManager, opts.ConveyorWithRetryWrapper)
	case addr.ArchiveAddress != nil:
		return NewArchiveStorage(NewArchiveStorageFileReader(addr.Path), NewArchiveStorageFileWriter(addr.Path))
	default:
		panic(fmt.Sprintf("invalid address given %#v", addr))
	}
}
