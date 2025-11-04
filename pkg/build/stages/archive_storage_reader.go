package stages

type ArchiveStorageReader interface {
}

// TODO should implement ArchiveStorageReader
type ArchiveStorageFileReader struct {
	Path string
}

func NewArchiveStorageFileReader(path string) *ArchiveStorageFileReader {
	return &ArchiveStorageFileReader{
		Path: path,
	}
}
