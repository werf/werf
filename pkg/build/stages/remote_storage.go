package stages

import (
	//TODO убрат
	"github.com/werf/werf/v2/pkg/deploy/bundles"
	"github.com/werf/werf/v2/pkg/docker_registry"
)

// TODO Should implement StorageAccessor
type RemoteStorage struct {
	RegistryAddress       *bundles.RegistryAddress
	StorageRegistryClient StorageRegistryClient
	RegistryClient        docker_registry.Interface
}

func NewRemoteStorage(registryAddress *bundles.RegistryAddress, storageRegistryClient StorageRegistryClient, registryClient docker_registry.Interface) *RemoteStorage {
	return &RemoteStorage{
		RegistryAddress:       registryAddress,
		StorageRegistryClient: storageRegistryClient,
		RegistryClient:        registryClient,
	}
}
