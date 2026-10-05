package stages

import (
	"context"
	"fmt"
	"strings"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/build"
	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/ref"
	"github.com/werf/werf/v2/pkg/storage"
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

			infoGetters, err = c.GetImageInfoGettersFromReport(ctx, image.InfoGetterOptions{OnlyFinal: false})
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

		var stageRefs []string
		for _, infoGetter := range infoGetters {
			logboek.Context(ctx).Default().LogFDetails("Copying stage: %s\n", infoGetter.Tag)

			reference, err := ref.ParseReference(infoGetter.Tag)
			if err != nil {
				return fmt.Errorf("unable to parse reference %q: %w", infoGetter.Tag, err)
			}

			reference.Repo = s.RegistryAddress.Repo
			reference.Tag = infoGetter.Tag

			infoGetterName := infoGetter.GetName()
			stageRefs = append(stageRefs, infoGetterName)

			if err = fromRemote.RegistryClient.CopyImage(ctx, infoGetterName, reference.FullName(), docker_registry.CopyImageOptions{}); err != nil {
				return fmt.Errorf("error copying stage %s into %s: %w", infoGetterName, reference.FullName(), err)
			}
		}

		return s.copyImportMetadata(ctx, fromRemote, opts.ProjectName, stageRefs)
	})
}

func (s *RemoteStorage) copyAllFromRemote(ctx context.Context, fromRemote *RemoteStorage, opts copyToOptions) error {
	stageIds, err := fromRemote.StorageManager.StagesStorage.GetStagesIDs(ctx, opts.ProjectName)
	if err != nil {
		return fmt.Errorf("unable to get stages: %w", err)
	}

	var stageRefs []string
	for _, stageId := range stageIds {
		logboek.Context(ctx).Default().LogFDetails("Copying stage: %s\n", stageId)

		stageDesc, err := fromRemote.StorageManager.StagesStorage.GetStageDesc(ctx, opts.ProjectName, stageId)
		if err != nil {
			return fmt.Errorf("unable to get description of stage %s: %w", stageId, err)
		}

		reference, err := ref.ParseReference(stageId.Digest)
		if err != nil {
			return fmt.Errorf("unable to parse stage %q reference: %w", stageId, err)
		}

		reference.Repo = s.RegistryAddress.Repo
		reference.Tag = stageDesc.Info.Tag

		stageName := stageDesc.Info.Name
		stageRefs = append(stageRefs, stageName)

		if err = fromRemote.RegistryClient.CopyImage(ctx, stageName, reference.FullName(), docker_registry.CopyImageOptions{}); err != nil {
			return fmt.Errorf("error copying stage %s into %s: %w", stageName, reference.FullName(), err)
		}
	}

	return s.copyImportMetadata(ctx, fromRemote, opts.ProjectName, stageRefs)
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
			return fmt.Errorf("unable to parse stage %q reference: %w", stageId, err)
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

func (s *RemoteStorage) copyImportMetadata(ctx context.Context, fromRemote *RemoteStorage, projectName string, stageRefs []string) error {
	metadataRefs, err := fromRemote.getImportMetadataImageRefs(ctx, projectName, stageRefs)
	if err != nil {
		return err
	}
	for _, sourceRef := range metadataRefs {
		_, tag := image.ParseRepositoryAndTag(sourceRef)
		destinationRef := fmt.Sprintf("%s:%s", s.RegistryAddress.Repo, tag)
		if err := fromRemote.RegistryClient.CopyImage(ctx, sourceRef, destinationRef, docker_registry.CopyImageOptions{}); err != nil {
			return fmt.Errorf("copy import metadata %s into %s: %w", sourceRef, destinationRef, err)
		}
	}
	return nil
}

func (s *RemoteStorage) getImportMetadataImageRefs(ctx context.Context, projectName string, stageRefs []string) ([]string, error) {
	checksums := make(map[string]bool)
	for _, stageRef := range stageRefs {
		info, err := s.RegistryClient.GetRepoImage(ctx, stageRef)
		if err != nil {
			return nil, fmt.Errorf("get copied stage %s import references: %w", stageRef, err)
		}
		infos := append([]*image.Info{info}, info.Index...)
		for _, platformInfo := range infos {
			for label, value := range platformInfo.Labels {
				if strings.HasPrefix(label, image.WerfImportChecksumLabelPrefix) && value != "" {
					checksums[value] = true
				}
			}
		}
	}
	if len(checksums) == 0 {
		return nil, nil
	}
	ids, err := s.StorageManager.StagesStorage.GetImportMetadataIDs(ctx, projectName)
	if err != nil {
		return nil, fmt.Errorf("get import metadata IDs: %w", err)
	}
	var refs []string
	for _, id := range ids {
		metadata, err := s.StorageManager.StagesStorage.GetImportMetadata(ctx, projectName, id)
		if storage.IsErrBrokenImage(err) || storage.IsErrImportMetadataNotFound(err) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("get import metadata %s: %w", id, err)
		}
		if checksums[metadata.Checksum] {
			refs = append(refs, fmt.Sprintf(storage.RepoImportMetadata_ImageNameFormat, s.RegistryAddress.Repo, id))
		}
	}
	return refs, nil
}
