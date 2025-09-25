package parallel

import (
	"io"
	"os"
)

type parallelWorkerReader struct {
	file   *os.File
	offset int64
}

func newParallelWorkerReader(f *os.File) *parallelWorkerReader {
	return &parallelWorkerReader{
		file:   f,
		offset: 0,
	}
}

func (r *parallelWorkerReader) NewSectionReader(offset, size int64) *io.SectionReader {
	r.offset += offset
	return io.NewSectionReader(r.file, r.offset, size)
}
