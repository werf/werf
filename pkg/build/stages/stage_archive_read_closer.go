package stages

import "io"

type StageArchiveReadCloser struct {
	reader io.Reader
	closer func() error
}

func NewStageArchiveReadCloser(reader io.Reader, closer func() error) *StageArchiveReadCloser {
	return &StageArchiveReadCloser{
		reader: reader,
		closer: closer,
	}
}

func (closer *StageArchiveReadCloser) Read(p []byte) (int, error) {
	return closer.reader.Read(p)
}

func (closer *StageArchiveReadCloser) Close() error {
	return closer.closer()
}
