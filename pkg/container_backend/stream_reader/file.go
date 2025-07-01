package stream_reader

import (
	"archive/tar"
	"io/fs"
)

type File struct {
	reader *tar.Reader
	header *tar.Header
}

func newFile(reader *tar.Reader, header *tar.Header) *File {
	return &File{reader: reader, header: header}
}

func (f *File) Path() string {
	return f.header.Name
}

func (f *File) Info() fs.FileInfo {
	return f.header.FileInfo()
}

func (f *File) Read(b []byte) (n int, err error) {
	return f.reader.Read(b)
}
