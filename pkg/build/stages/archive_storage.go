package stages

import (
	"context"
	"fmt"
	"io"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/build"
	"github.com/werf/werf/v2/pkg/image"
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

			if err := writer.WriteStageArchive(tag, func(w io.Writer) error {
				return fromRemote.RegistryClient.PullImageArchive(ctx, w, stageRef)
			}); err != nil {
				return fmt.Errorf("error copying stage %q: %w", stageRef, err)
			}
		}

		return nil
	})
}

func (s *ArchiveStorage) copyCurrentBuildFromRemote(ctx context.Context, fromRemote *RemoteStorage, opts copyToOptions) error {
	return s.Writer.WithTask(ctx, func(writer ArchiveStorageWriter) error {
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

				stageRef := infoGetter.GetName()
				if err := writer.WriteStageArchive(infoGetter.Tag, func(w io.Writer) error {
					return fromRemote.RegistryClient.PullImageArchive(ctx, w, stageRef)
				}); err != nil {
					return fmt.Errorf("error copying stage %q: %w", stageRef, err)
				}
			}

			return nil
		})
	})
}
