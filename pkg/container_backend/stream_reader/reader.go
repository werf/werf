package stream_reader

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// NewFileSystemStreamReader takes an image tar reader and returns file system stream reader.
func NewFileSystemStreamReader(imgTarReader *tar.Reader) (*FileSystemStreamReader, error) {
	fsReader, err := findRoot(imgTarReader)
	if err != nil {
		return nil, err
	}
	if fsReader == nil {
		return &FileSystemStreamReader{nil}, nil
	}
	return &FileSystemStreamReader{fsReader}, nil
}

// findRoot finds the root of file system inside of image tarball
func findRoot(imgTarReader *tar.Reader) (*tar.Reader, error) {
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

		switch http.DetectContentType(firstBytes) {
		case "application/x-tar", "application/octet-stream":
			bufLayers, err := bufferLayers(firstBytes, imgTarReader)
			if err != nil {
				return nil, err
			}
			return tar.NewReader(bufLayers), nil

		case "application/x-gzip":
			bufLayers, err := bufferLayers(firstBytes, imgTarReader)
			if err != nil {
				return nil, err
			}
			gzipReader, err := gzip.NewReader(bufLayers)
			if err != nil {
				return nil, fmt.Errorf("unable to create gzip reader: %w", err)
			}

			return tar.NewReader(gzipReader), nil
		default:
			continue // unsupported content type
		}
	}
}

// bufferLayers creates new buffer from image layers
func bufferLayers(firstBytes []byte, imgTarReader *tar.Reader) (*bytes.Buffer, error) {
	// don't forget to write the first bytes
	bufLayers := bytes.NewBuffer(firstBytes)

	// copy from second byte upt to the rest bytes
	if _, err := io.Copy(bufLayers, imgTarReader); err != nil {
		return nil, err
	}

	return bufLayers, nil
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
