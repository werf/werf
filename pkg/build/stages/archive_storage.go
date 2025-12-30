package stages

import (
	"context"
	"fmt"
	"io"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/build"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/util/stream"
)

const (
	tarGzExtension    = ".tar.gz"
	stagePathTemplate = "stages/%s" + tarGzExtension
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

func (s *ArchiveStorage) ReadStagesTags(ctx context.Context) ([]string, error) {
	stages, err := s.Reader.ReadStagesTags(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to read stages archive: %w", err)
	}

	return stages, nil
}

func (s *ArchiveStorage) GetStageArchiveOpener(stageTag string) *StageArchiveOpener {
	return NewStageArchiveOpener(s, stageTag)
}

func (s *ArchiveStorage) copyAllFromRemote(ctx context.Context, fromRemote *RemoteStorage, opts copyToOptions) error {
	stageIds, err := fromRemote.StorageManager.StagesStorage.GetStagesIDs(ctx, opts.ProjectName)
	if err != nil {
		return fmt.Errorf("unable to get stages: %w", err)
	}

	return s.Writer.WithTask(ctx, func(writer ArchiveStorageWriter) error {
		for _, stageId := range stageIds {
			logboek.Context(ctx).Default().LogFDetails("Copying stage: %s\n", stageId)

			stageDesc, err := fromRemote.StorageManager.StagesStorage.GetStageDesc(ctx, opts.ProjectName, stageId)
			if err != nil {
				return err
			}

			stageRef := stageDesc.Info.Name
			tag := stageDesc.Info.Tag

			producer := func(ctx context.Context, w io.Writer) error {
				if err := fromRemote.RegistryClient.PullImageArchive(ctx, w, stageRef); err != nil {
					return fmt.Errorf("error pulling stage %q archive: %w", stageRef, err)
				}

				return nil
			}

			consumer := func(ctx context.Context, r io.Reader) error {
				if err := writer.WriteStageArchive(tag, r); err != nil {
					return fmt.Errorf("error writing image %q into archive: %w", stageRef, err)
				}

				return nil
			}

			if err := stream.PipeProducerConsumer(ctx, producer, consumer); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *ArchiveStorage) copyCurrentBuildFromRemote(ctx context.Context, fromRemote *RemoteStorage, opts copyToOptions) error {
	return s.Writer.WithTask(ctx, func(writer ArchiveStorageWriter) error {
		return fromRemote.ConveyorWithRetry.WithRetryBlock(ctx, func(c *build.Conveyor) error {
			if _, err := c.Build(ctx, opts.BuildOptions); err != nil {
				return err
			}

			infoGetters, err := c.GetImageInfoGettersWithOpts(image.InfoGetterOptions{OnlyFinal: false})
			if err != nil {
				return err
			}

			for _, infoGetter := range infoGetters {
				logboek.Context(ctx).Default().LogFDetails("Copying stage: %s\n", infoGetter.Tag)

				name := infoGetter.GetName()
				tag := infoGetter.Tag

				producer := func(ctx context.Context, w io.Writer) error {
					if err := fromRemote.RegistryClient.PullImageArchive(ctx, w, name); err != nil {
						return fmt.Errorf("error pulling image %q archive: %w", name, err)
					}

					return nil
				}

				consumer := func(ctx context.Context, r io.Reader) error {
					if err := writer.WriteStageArchive(tag, r); err != nil {
						return fmt.Errorf("error writing image %q into archive: %w", name, err)
					}

					return nil
				}

				if err := stream.PipeProducerConsumer(ctx, producer, consumer); err != nil {
					return err
				}
			}

			return nil
		})
	})
}
