package bundles

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
)

type BundleArchiveReader interface {
	String() string
	ReadChartArchive() ([]byte, error)
	ReadImageArchive(imageTag string) (*ImageArchiveReadCloser, error)
}

type BundleArchiveFileReader struct {
	Path string
}

func NewBundleArchiveFileReader(path string) *BundleArchiveFileReader {
	return &BundleArchiveFileReader{Path: path}
}

func (reader *BundleArchiveFileReader) String() string {
	return reader.Path
}

func (reader *BundleArchiveFileReader) ReadChartArchive() ([]byte, error) {
	treader, closer, err := reader.openForReading()
	defer closer()

	if err != nil {
		return nil, fmt.Errorf("unable to open bundle archive: %w", err)
	}

	b := bytes.NewBuffer(nil)

	for {
		header, err := treader.Next()
		if err == io.EOF {
			return nil, fmt.Errorf("no chart archive found in the bundle archive %q", reader.Path)
		}
		if err != nil {
			return nil, fmt.Errorf("error reading tar archive: %w", err)
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}
		if header.Name != chartArchiveFileName {
			continue
		}

		if _, err := io.Copy(b, treader); err != nil {
			return nil, fmt.Errorf("unable to read chart archive %q from the bundle archive %q: %w", chartArchiveFileName, reader.Path, err)
		}

		return b.Bytes(), nil
	}
}

func (reader *BundleArchiveFileReader) ReadImageArchive(imageTag string) (*ImageArchiveReadCloser, error) {
	treader, closer, err := reader.openForReading()
	if err != nil {
		defer closer()
		return nil, fmt.Errorf("unable to open bundle archive: %w", err)
	}

	for {
		header, err := treader.Next()
		if err == io.EOF {
			return nil, fmt.Errorf("no image tag %q found in the bundle archive %q", imageTag, reader.Path)
		}
		if err != nil {
			return nil, fmt.Errorf("error reading tar archive: %w", err)
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}

		if header.Name == fmt.Sprintf("images/%s.tar.gz", imageTag) {
			unzipper, err := gzip.NewReader(treader)
			if err != nil {
				return nil, fmt.Errorf("unable to create gzip reader for image archive: %w", err)
			}

			return NewImageArchiveReadCloser(unzipper, func() error {
				if err := unzipper.Close(); err != nil {
					return fmt.Errorf("unable to close gzip reader for image archive: %w", err)
				}
				return closer()
			}), nil
		}
	}
}

func (reader *BundleArchiveFileReader) openForReading() (*tar.Reader, func() error, error) {
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
