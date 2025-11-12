package stages

import (
	"context"
	"fmt"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/build"
	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/ref"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/storage/manager"
)

type RemoteStorage struct {
	RegistryAddress   *ref.RegistryAddress
	RegistryClient    docker_registry.Interface
	StorageManager    *manager.StorageManager
	ConveyorWithRetry *build.ConveyorWithRetryWrapper

	AllStages bool
}

func NewRemoteStorage(addr *ref.RegistryAddress, dockerRegistry docker_registry.Interface, stagesStorage *manager.StorageManager, conveyorWithRetry *build.ConveyorWithRetryWrapper, allStages bool) *RemoteStorage {
	return &RemoteStorage{
		RegistryAddress:   addr,
		RegistryClient:    dockerRegistry,
		StorageManager:    stagesStorage,
		ConveyorWithRetry: conveyorWithRetry,
		AllStages:         allStages,
	}
}

func (s *RemoteStorage) CopyTo(ctx context.Context, to StorageAccessor, opts copyToOptions) error {
	return to.CopyFromRemote(ctx, s, opts)
}

func (s *RemoteStorage) CopyFromArchive(ctx context.Context, fromArchive *ArchiveStorage, opts copyToOptions) error {
	return s.copyAllFromArchive(ctx, fromArchive)
}

func (s *RemoteStorage) CopyFromRemote(ctx context.Context, fromRemote *RemoteStorage, opts copyToOptions) error {
	if s.AllStages {
		return s.copyAllFromRemote(ctx, fromRemote, opts.ProjectName)
	}
	return s.copyCurrentBuildStagesFromRemote(ctx, fromRemote, opts.ProjectName)
}

func (s *RemoteStorage) copyCurrentBuildStagesFromRemote(ctx context.Context, fromRemote *RemoteStorage, projectName string, opts ...storage.Option) error {
	panic("not implemented yet")
}

func (s *RemoteStorage) copyAllFromRemote(ctx context.Context, fromRemote *RemoteStorage, projectName string, opts ...storage.Option) error {
	stageIds, err := fromRemote.StorageManager.StagesStorage.GetStagesIDs(ctx, projectName, opts...)
	if err != nil {
		return fmt.Errorf("unable to get stages: %w", err)
	}

	if err := logboek.Context(ctx).LogProcess("Copy stages from container registry").DoError(func() error {
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

			logboek.Context(ctx).Default().LogFDetails("Source: %s\n", stageDesc.Info.Name)
			logboek.Context(ctx).Default().LogFDetails("Destination: %s\n", reference.FullName())

			if err = fromRemote.RegistryClient.CopyImage(ctx, stageDesc.Info.Name, reference.FullName(), docker_registry.CopyImageOptions{}); err != nil {
				return fmt.Errorf("error copying stage %s into %s: %w", stageDesc.Info.Name, reference.FullName(), err)
			}
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (s *RemoteStorage) copyAllFromArchive(ctx context.Context, fromArchive *ArchiveStorage) error {
	panic("not implemented yet")
}
