package stages

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	tarGzExtension = ".tar.gz"
)

type ArchiveStorageReader interface {
	String() string
	ReadStagesTags() ([]string, error)
	ReadArchiveStage(stageTag string) (*StageArchiveReadCloser, error)
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

func (reader *ArchiveStorageFileReader) ReadStagesTags() ([]string, error) {
	treader, closer, err := reader.openForReading()
	if err != nil {
		return nil, fmt.Errorf("error opening archive: %v", err)
	}
	defer closer()

	var tags []string

	for {
		header, err := treader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading tar archive: %v", err)
		}

		if header.Typeflag == tar.TypeReg {
			filename := filepath.Base(header.Name)
			if strings.HasSuffix(filename, tarGzExtension) {
				nameWithoutExt := strings.TrimSuffix(filename, tarGzExtension)
				tags = append(tags, nameWithoutExt)
			}
		}
	}

	return tags, nil
}

func (reader *ArchiveStorageFileReader) ReadArchiveStage(stageTag string) (*StageArchiveReadCloser, error) {
	treader, closer, err := reader.openForReading()
	if err != nil {
		return nil, fmt.Errorf("unable to open stages archive: %w", err)
	}
	defer closer()

	for {
		header, err := treader.Next()
		if err == io.EOF {
			return nil, fmt.Errorf("no stage tag %q found in the stages archive %q", stageTag, reader.Path)
		}
		if err != nil {
			return nil, fmt.Errorf("error reading tar archive: %w", err)
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}

		if header.Name == fmt.Sprintf("stages/%s.tar.gz", stageTag) {
			unzipper, err := gzip.NewReader(treader)
			if err != nil {
				return nil, fmt.Errorf("unable to create gzip reader for stages archive: %w", err)
			}

			return NewStageArchiveReadCloser(unzipper, func() error {
				if err := unzipper.Close(); err != nil {
					return fmt.Errorf("unable to close gzip reader for stage archive: %w", err)
				}

				return closer()
			}), nil
		}
	}
}

func (reader *ArchiveStorageFileReader) openForReading() (*tar.Reader, func() error, error) {
	f, err := os.Open(reader.Path)
	if err != nil {
		return nil, func() error { return nil }, err
	}

	unzipper, err := gzip.NewReader(f)
	if err != nil {
		return nil, f.Close, fmt.Errorf("unable to open stages archive gzip %q: %w", reader.Path, err)
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
