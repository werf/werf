package stages

import (
	"bytes"
	"context"
	"fmt"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/storage"
)

const (
	archiveStageFileName = "stage.tar.gz"
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
	return s.copyAllFromRemote(ctx, fromRemote, opts.ProjectName)
}

func (s *ArchiveStorage) copyAllFromRemote(ctx context.Context, fromRemote *RemoteStorage, projectName string, opts ...storage.Option) error {
	stageIds, err := fromRemote.StorageManager.StagesStorage.GetStagesIDs(ctx, projectName, opts...)
	if err != nil {
		return fmt.Errorf("unable to get stages: %w", err)
	}

	if err := s.Writer.Open(); err != nil {
		return fmt.Errorf("unable to open target stages archive: %w", err)
	}

	if err := logboek.Context(ctx).LogProcess("Saving stages into archive").DoError(func() error {
		for _, stageId := range stageIds {
			stageDesc, err := fromRemote.StorageManager.StagesStorage.GetStageDesc(ctx, projectName, stageId)
			if err != nil {
				return err
			}

			stageRef := stageDesc.Info.Name
			tag := stageDesc.Info.Tag

			logboek.Context(ctx).Default().LogFDetails("Saving stage %s\n", stageRef)

			stageBytes := bytes.NewBuffer(nil)

			if err := fromRemote.RegistryClient.PullImageArchive(ctx, stageBytes, stageRef); err != nil {
				return fmt.Errorf("error pulling stage %q archive: %w", stageRef, err)
			}

			if err := s.Writer.WriteStageArchive(tag, stageBytes.Bytes()); err != nil {
				return fmt.Errorf("error writing image %q into bundle archive: %w", stageRef, err)
			}
		}

		return nil
	}); err != nil {
		return err
	}

	if err := s.Writer.Save(); err != nil {
		return fmt.Errorf("error saving destination bundle archive: %w", err)
	}

	return nil
}
