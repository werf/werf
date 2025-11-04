package stages

import (
	"fmt"

	"github.com/werf/werf/v2/pkg/deploy/bundles"
	"github.com/werf/werf/v2/pkg/docker_registry"
)

type copyToOptions struct {
}

type StorageAccessor interface {
}

type StorageAccessorOptions struct {
	StorageRegistryClient StorageRegistryClient
	RegistryClient        docker_registry.Interface
}

func NewStorageAccessor(addr *bundles.Addr, opts StorageAccessorOptions) StorageAccessor {
	switch {
	case addr.RegistryAddress != nil:
		return NewRemoteStorage(addr.RegistryAddress, opts.StorageRegistryClient, opts.RegistryClient)
	case addr.ArchiveAddress != nil:
		return NewArchiveStorage(NewArchiveStorageFileReader(addr.ArchiveAddress.Path), NewArchiveStorageFileWriter(addr.ArchiveAddress.Path))
	default:
		panic(fmt.Sprintf("invalid address given %#v", addr))
	}
}
