package stages

import (
	"context"
	"fmt"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/build"
	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/ref"
	"github.com/werf/werf/v2/pkg/storage/manager"
)

type RemoteStorage struct {
	RegistryAddress   *ref.RegistryAddress
	RegistryClient    docker_registry.Interface
	StorageManager    *manager.StorageManager
	ConveyorWithRetry *build.ConveyorWithRetryWrapper
}

func NewRemoteStorage(addr *ref.RegistryAddress, dockerRegistry docker_registry.Interface, storageManager *manager.StorageManager, conveyorWithRetry *build.ConveyorWithRetryWrapper) *RemoteStorage {
	return &RemoteStorage{
		RegistryAddress:   addr,
		RegistryClient:    dockerRegistry,
		StorageManager:    storageManager,
		ConveyorWithRetry: conveyorWithRetry,
	}
}

func (s *RemoteStorage) CopyTo(ctx context.Context, to StorageAccessor, opts copyToOptions) error {
	return to.CopyFromRemote(ctx, s, opts)
}

func (s *RemoteStorage) CopyFromArchive(ctx context.Context, fromArchive *ArchiveStorage, opts copyToOptions) error {
	return s.copyFromArchive(ctx, fromArchive)
}

func (s *RemoteStorage) CopyFromRemote(ctx context.Context, fromRemote *RemoteStorage, opts copyToOptions) error {
	if opts.All {
		return s.copyAllFromRemote(ctx, fromRemote, opts)
	}

	return s.copyCurrentBuildStagesFromRemote(ctx, fromRemote, opts)
}

func (s *RemoteStorage) copyCurrentBuildStagesFromRemote(ctx context.Context, fromRemote *RemoteStorage, opts copyToOptions) error {
	return fromRemote.ConveyorWithRetry.WithRetryBlock(ctx, func(c *build.Conveyor) error {
		var infoGetters []*image.InfoGetter
		var err error

		if c.UseBuildReport {
			logboek.Context(ctx).Debug().LogFDetails("Avoid building because of using build report: %s\n", c.BuildReportPath)

			infoGetters, err = c.GetImageInfoGettersFromReport(image.InfoGetterOptions{OnlyFinal: false})
			if err != nil {
				return fmt.Errorf("unable to get image info getters from build report: %w", err)
			}
		} else {
			if _, err := c.Build(ctx, opts.BuildOptions); err != nil {
				return fmt.Errorf("error while building: %w", err)
			}

			infoGetters, err = c.GetImageInfoGettersWithOpts(image.InfoGetterOptions{OnlyFinal: false})
			if err != nil {
				return fmt.Errorf("unable to get image info getters: %w", err)
			}
		}

		for _, infoGetter := range infoGetters {
			logboek.Context(ctx).Default().LogFDetails("Copying stage: %s\n", infoGetter.Tag)

			reference, err := ref.ParseReference(infoGetter.Tag)
			if err != nil {
				return err
			}

			reference.Repo = s.RegistryAddress.Repo
			reference.Tag = infoGetter.Tag

			infoGetterName := infoGetter.GetName()

			if err = fromRemote.RegistryClient.CopyImage(ctx, infoGetterName, reference.FullName(), docker_registry.CopyImageOptions{}); err != nil {
				return fmt.Errorf("error copying stage %s into %s: %w", infoGetterName, reference.FullName(), err)
			}
		}

		return nil
	})
}

func (s *RemoteStorage) copyAllFromRemote(ctx context.Context, fromRemote *RemoteStorage, opts copyToOptions) error {
	stageIds, err := fromRemote.StorageManager.StagesStorage.GetStagesIDs(ctx, opts.ProjectName)
	if err != nil {
		return fmt.Errorf("unable to get stages: %w", err)
	}

	for _, stageId := range stageIds {
		logboek.Context(ctx).Default().LogFDetails("Copying stage: %s\n", stageId)

		stageDesc, err := fromRemote.StorageManager.StagesStorage.GetStageDesc(ctx, opts.ProjectName, stageId)
		if err != nil {
			return err
		}

		reference, err := ref.ParseReference(stageId.Digest)
		if err != nil {
			return err
		}

		reference.Repo = s.RegistryAddress.Repo
		reference.Tag = stageDesc.Info.Tag

		stageName := stageDesc.Info.Name

		if err = fromRemote.RegistryClient.CopyImage(ctx, stageName, reference.FullName(), docker_registry.CopyImageOptions{}); err != nil {
			return fmt.Errorf("error copying stage %s into %s: %w", stageName, reference.FullName(), err)
		}
	}

	return nil
}

func (s *RemoteStorage) copyFromArchive(ctx context.Context, fromArchive *ArchiveStorage) error {
	stageIds, err := fromArchive.ReadStagesTags(ctx)
	if err != nil {
		return fmt.Errorf("error reading stages: %w", err)
	}

	for _, stageId := range stageIds {
		logboek.Context(ctx).Default().LogFDetails("Copying stage: %s\n", stageId)

		reference, err := ref.ParseReference(stageId)
		if err != nil {
			return err
		}

		reference.Repo = s.RegistryAddress.Repo
		reference.Tag = stageId

		stageArchiveOpener := fromArchive.GetStageArchiveOpener(stageId)
		stageArchiveOpener.SetContext(ctx)

		if err := s.RegistryClient.PushImageArchive(ctx, stageArchiveOpener, reference.FullName()); err != nil {
			return fmt.Errorf("error copying stage %q archive: %w", stageId, err)
		}
	}

	return nil
}
