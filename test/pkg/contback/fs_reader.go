package contback

import (
	"archive/tar"
	"bytes"
	"errors"
	"io"
	"strings"

	. "github.com/onsi/gomega"
)

// newFileSystemReader takes an image tar reader and returns file system tar reader.
func newFileSystemReader(imgTarReader *tar.Reader) (*FileSystemReader, error) {
	for {
		header, err := imgTarReader.Next()

		if errors.Is(err, io.EOF) {
			return &FileSystemReader{tar.NewReader(bytes.NewBuffer([]byte{}))}, nil
		}
		if err != nil {
			return nil, err
		}

		if !header.FileInfo().Mode().IsRegular() {
			continue
		}

		if !strings.HasPrefix(header.Name, "blobs/sha256/") {
			continue
		}

		// test "is json file"
		firstByte := make([]byte, 1)
		if _, err = imgTarReader.Read(firstByte); err != nil {
			return nil, err
		}
		if string(firstByte) == "{" {
			continue
		}

		// don't forget to write the first byte
		bufFs := bytes.NewBuffer(firstByte)
		// copy from second byte upt to the last byte
		if _, err = io.Copy(bufFs, imgTarReader); err != nil {
			return nil, err
		}

		return &FileSystemReader{tar.NewReader(bufFs)}, nil
	}
}

type FileSystemReader struct {
	reader *tar.Reader
}

func (r *FileSystemReader) Next() *tar.Header {
	hdr, err := r.reader.Next()
	if err != io.EOF {
		Expect(err).To(Succeed())
	}
	return hdr
}
