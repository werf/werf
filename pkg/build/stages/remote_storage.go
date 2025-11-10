package stages

import (
	"context"

	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/ref"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/storage/manager"
)

type RemoteStorage struct {
	RegistryAddress *ref.RegistryAddress
	RegistryClient  docker_registry.Interface
	StorageManager  *manager.StorageManager
}

func NewRemoteStorage(addr *ref.RegistryAddress, stagesStorage *manager.StorageManager, dockerRegistry docker_registry.Interface) *RemoteStorage {
	return &RemoteStorage{
		RegistryAddress: addr,
		RegistryClient:  dockerRegistry,
		StorageManager:  stagesStorage,
	}
}

func (s *RemoteStorage) CopyTo(ctx context.Context, to StorageAccessor, opts copyToOptions) error {
	return to.CopyFromRemote(ctx, s, opts)
}

func (s *RemoteStorage) CopyFromArchive(ctx context.Context, fromArchive *ArchiveStorage, opts copyToOptions) error {
	return s.copyAllFromArchive(ctx, fromArchive, opts)
}

func (s *RemoteStorage) CopyFromRemote(ctx context.Context, fromRemote *RemoteStorage, opts copyToOptions) error {
	return s.copyAllFromRemote(ctx, fromRemote, opts.ProjectName)
}

func (s *RemoteStorage) copyAllFromRemote(ctx context.Context, fromRemote *RemoteStorage, projectName string, opts ...storage.Option) error {
	stageIds, err := fromRemote.StorageManager.StagesStorage.GetStagesIDs(ctx, projectName, opts...)
	if err != nil {
		return err
	}

	for _, stageId := range stageIds {
		stageDesc, err := fromRemote.StorageManager.StagesStorage.GetStageDesc(ctx, projectName, stageId)
		if err != nil {
			return err
		}

		reference, err := ref.ParseReference(stageId.Digest)
		if err != nil {
			return err
		}
		reference.Repo = s.RegistryAddress.Repo
		reference.Tag = stageDesc.Info.Tag

		if err = fromRemote.RegistryClient.CopyImage(ctx, stageDesc.Info.Name, reference.FullName(), docker_registry.CopyImageOptions{}); err != nil {
			return err
		}
	}
	return nil
}
