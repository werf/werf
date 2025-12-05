package stages

import "io"

type ArchiveStageReadCloser struct {
	reader io.Reader
	closer func() error
}

func NewStageArchiveReadCloser(reader io.Reader, closer func() error) *ArchiveStageReadCloser {
	return &ArchiveStageReadCloser{
		reader: reader,
		closer: closer,
	}
}

func (closer *ArchiveStageReadCloser) Read(p []byte) (int, error) {
	return closer.reader.Read(p)
}

func (closer *ArchiveStageReadCloser) Close() error {
	return closer.closer()
}
