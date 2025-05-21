package contback

import (
	"archive/tar"
	"bytes"

	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/container_backend/stream_reader"
)

type fileSystemStreamReaderWrapper struct {
	reader *stream_reader.FileSystemStreamReader
}

func NewFileSystemReaderWrapper(imgReader *bytes.Reader) *fileSystemStreamReaderWrapper {
	fsReader, err := stream_reader.NewFileSystemStreamReader(tar.NewReader(imgReader))
	Expect(err).To(Succeed())

	return &fileSystemStreamReaderWrapper{reader: fsReader}
}

func (w *fileSystemStreamReaderWrapper) Next() *stream_reader.File {
	f, err := w.reader.Next()
	Expect(err).To(Succeed())
	return f
}
