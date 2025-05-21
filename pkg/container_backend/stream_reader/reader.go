package stream_reader

import (
	"archive/tar"
	"bytes"
	"errors"
	"io"
	"strings"
)

// NewFileSystemStreamReader takes an image tar reader and returns file system stream reader.
func NewFileSystemStreamReader(imgTarReader *tar.Reader) (*FileSystemStreamReader, error) {
	fsRoot, err := findRoot(imgTarReader)
	if err != nil {
		return nil, err
	}
	if fsRoot == nil {
		return &FileSystemStreamReader{nil}, nil
	}
	return &FileSystemStreamReader{tar.NewReader(fsRoot)}, nil
}

func findRoot(imgTarReader *tar.Reader) (*bytes.Buffer, error) {
	for {
		header, err := imgTarReader.Next()

		if errors.Is(err, io.EOF) {
			return nil, nil
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

		return bufFs, nil
	}
}

type FileSystemStreamReader struct {
	reader *tar.Reader
}

func (r *FileSystemStreamReader) Next() (*File, error) {
	if r.reader == nil {
		return nil, nil
	}

	header, err := r.reader.Next()

	if errors.Is(err, io.EOF) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return newFile(r.reader, header), nil
}
