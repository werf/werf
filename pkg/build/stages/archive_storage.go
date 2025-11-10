package stages

import "context"

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

func (storage *ArchiveStorage) CopyTo(ctx context.Context, to StorageAccessor, opts copyToOptions) error {
	return to.CopyFromArchive(ctx, storage, opts)
}

func (storage *ArchiveStorage) CopyFromArchive(ctx context.Context, fromArchive *ArchiveStorage, opts copyToOptions) error {
	return nil
}

func (storage *ArchiveStorage) CopyFromRemote(ctx context.Context, fromRemote *RemoteStorage, opts copyToOptions) error {
	return nil
}
