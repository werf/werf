package stream_reader

import (
	"archive/tar"
	"bytes"
	"errors"
	"io"
	"net/http"
	"slices"
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

		// Look for the file system tar archive inside of image tar archive
		firstBytes := make([]byte, min(header.Size, 512))
		if _, err = imgTarReader.Read(firstBytes); err != nil && !errors.Is(err, io.EOF) {
			return nil, err
		}

		detectedContentType := http.DetectContentType(firstBytes)
		if !slices.Contains([]string{"application/x-tar", "application/octet-stream"}, detectedContentType) {
			continue
		}

		// don't forget to write the first bytes
		bufFs := bytes.NewBuffer(firstBytes)
		// copy from second byte upt to the rest bytes
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

func (r *FileSystemStreamReader) Find(predicate func(file *File) bool) (*File, bool, error) {
	for {
		f, err := r.Next()
		if err != nil {
			return nil, false, err
		}
		if f == nil {
			return nil, false, nil
		}
		if predicate(f) {
			return f, true, nil
		}
	}
}
