package stages

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
)

type ArchiveStorageReader interface {
	String() string
	ReadArchiveStage() ([]byte, error)
}

type ArchiveStorageFileReader struct {
	Path string
}

func NewArchiveStorageFileReader(path string) *ArchiveStorageFileReader {
	return &ArchiveStorageFileReader{
		Path: path,
	}
}

func (reader *ArchiveStorageFileReader) String() string {
	return reader.Path
}

func (reader *ArchiveStorageFileReader) ReadArchiveStage() ([]byte, error) {
	treader, closer, err := reader.openForReading()
	defer closer()

	if err != nil {
		return nil, fmt.Errorf("unable to open stages archive: %w", err)
	}

	b := bytes.NewBuffer(nil)

	for {
		header, err := treader.Next()
		if err == io.EOF {
			return nil, fmt.Errorf("no stages archive found in the archive %q", reader.Path)
		}
		if err != nil {
			return nil, fmt.Errorf("error reading tar archive: %w", err)
		}

		if header.Typeflag == tar.TypeReg {
			continue
		}

		if header.Name != archiveStageFileName {
			continue
		}

		if _, err := io.Copy(b, treader); err != nil {
			return nil, fmt.Errorf("unable to read stage archive %q from the archive %q: %w", archiveStageFileName, reader.Path, err)
		}

		return b.Bytes(), nil
	}
}

func (reader *ArchiveStorageFileReader) openForReading() (*tar.Reader, func() error, error) {
	f, err := os.Open(reader.Path)
	if err != nil {
		return nil, func() error { return nil }, err
	}

	unzipper, err := gzip.NewReader(f)
	if err != nil {
		return nil, f.Close, fmt.Errorf("unable to open bundle archive gzip %q: %w", reader.Path, err)
	}

	closer := func() error {
		if err := unzipper.Close(); err != nil {
			return fmt.Errorf("unable to close gzipper for %q: %w", reader.Path, err)
		}
		if err := f.Close(); err != nil {
			return fmt.Errorf("unable to close %q: %w", reader.Path, err)
		}
		return nil
	}

	return tar.NewReader(unzipper), closer, nil
}
