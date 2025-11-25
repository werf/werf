package stages

import (
	"bytes"
	"context"
	"fmt"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/build"
	"github.com/werf/werf/v2/pkg/image"
)

type ArchiveStorage struct {
	Reader ArchiveStorageReader
	Writer ArchiveStorageWriter
}

func NewArchiveStorage(reader ArchiveStorageReader, writer ArchiveStorageWriter) *ArchiveStorage {
	return &ArchiveStorage{
		Reader: reader,
		Writer: writer,
	}
}

func (s *ArchiveStorage) CopyTo(ctx context.Context, to StorageAccessor, opts copyToOptions) error {
	return to.CopyFromArchive(ctx, s, opts)
}

func (s *ArchiveStorage) CopyFromArchive(ctx context.Context, fromArchive *ArchiveStorage, opts copyToOptions) error {
	panic("not implemented yet")
}

func (s *ArchiveStorage) CopyFromRemote(ctx context.Context, fromRemote *RemoteStorage, opts copyToOptions) error {
	if opts.All {
		return s.copyAllFromRemote(ctx, fromRemote, opts)
	}

	return s.copyCurrentBuildFromRemote(ctx, fromRemote, opts)
}

func (s *ArchiveStorage) ReadStagesTags() ([]string, error) {
	stages, err := s.Reader.ReadStagesTags()
	if err != nil {
		return nil, fmt.Errorf("unable to read stages archive: %w", err)
	}

	return stages, nil
}

func (s *ArchiveStorage) GetStageArchiveOpener(stageTag string) *StageArchiveOpener {
	return NewStageArchiveOpener(s, stageTag)
}

func (s *ArchiveStorage) copyAllFromRemoteDeprecated(ctx context.Context, fromRemote *RemoteStorage, opts copyToOptions) error {
	stageIds, err := fromRemote.StorageManager.StagesStorage.GetStagesIDs(ctx, opts.ProjectName)
	if err != nil {
		return fmt.Errorf("unable to get stages: %w", err)
	}

	if err := s.Writer.Open(); err != nil {
		return fmt.Errorf("unable to open target stages archive: %w", err)
	}

	for _, stageId := range stageIds {
		logboek.Context(ctx).Default().LogFDetails("Copying stage: %s\n", stageId)

		stageDesc, err := fromRemote.StorageManager.StagesStorage.GetStageDesc(ctx, opts.ProjectName, stageId)
		if err != nil {
			return err
		}

		stageRef := stageDesc.Info.Name
		tag := stageDesc.Info.Tag

		stageBytes := bytes.NewBuffer(nil)

		if err := fromRemote.RegistryClient.PullImageArchive(ctx, stageBytes, stageRef); err != nil {
			return fmt.Errorf("error pulling stage %q archive: %w", stageRef, err)
		}

		if err := s.Writer.WriteStageArchive(tag, stageBytes.Bytes()); err != nil {
			return fmt.Errorf("error writing image %q into bundle archive: %w", stageRef, err)
		}
	}

	if err := s.Writer.Save(); err != nil {
		return fmt.Errorf("error saving destination bundle archive: %w", err)
	}

	return nil
}

func (s *ArchiveStorage) copyAllFromRemote(ctx context.Context, fromRemote *RemoteStorage, opts copyToOptions) error {
	stageIds, err := fromRemote.StorageManager.StagesStorage.GetStagesIDs(ctx, opts.ProjectName)
	if err != nil {
		return fmt.Errorf("unable to get stages: %w", err)
	}

	return s.Writer.WithTask(func(writer ArchiveStorageWriter) error {
		for _, stageId := range stageIds {
			logboek.Context(ctx).Default().LogFDetails("Copying stage: %s\n", stageId)

			stageDesc, err := fromRemote.StorageManager.StagesStorage.GetStageDesc(ctx, opts.ProjectName, stageId)
			if err != nil {
				return err
			}

			stageRef := stageDesc.Info.Name
			tag := stageDesc.Info.Tag

			stageBytes := bytes.NewBuffer(nil)

			if err := fromRemote.RegistryClient.PullImageArchive(ctx, stageBytes, stageRef); err != nil {
				return fmt.Errorf("error pulling stage %q archive: %w", stageRef, err)
			}

			if err := writer.WriteStageArchive(tag, stageBytes.Bytes()); err != nil {
				return fmt.Errorf("error writing image %q into bundle archive: %w", stageRef, err)
			}
		}
		return nil
	})
}

func (s *ArchiveStorage) copyCurrentBuildFromRemote(ctx context.Context, fromRemote *RemoteStorage, opts copyToOptions) error {
	if err := s.Writer.Open(); err != nil {
		return fmt.Errorf("unable to open target stages archive: %w", err)
	}

	return fromRemote.ConveyorWithRetry.WithRetryBlock(ctx, func(c *build.Conveyor) error {
		if _, err := c.Build(ctx, opts.BuildOptions); err != nil {
			return err
		}

		infoGetters, err := c.GetImageInfoGetters(image.InfoGetterOptions{OnlyFinal: false})
		if err != nil {
			return err
		}

		for _, infoGetter := range infoGetters {
			logboek.Context(ctx).Default().LogFDetails("Copying stage: %s\n", infoGetter.Tag)

			stageBytes := bytes.NewBuffer(nil)

			if err := fromRemote.RegistryClient.PullImageArchive(ctx, stageBytes, infoGetter.GetName()); err != nil {
				return fmt.Errorf("error pulling stage %q archive: %w", infoGetter.GetName(), err)
			}

			if err := s.Writer.WriteStageArchive(infoGetter.Tag, stageBytes.Bytes()); err != nil {
				return fmt.Errorf("error writing image %q into bundle archive: %w", infoGetter.GetName(), err)
			}
		}

		if err := s.Writer.Save(); err != nil {
			return fmt.Errorf("error saving destination bundle archive: %w", err)
		}

		return nil
	})
}
