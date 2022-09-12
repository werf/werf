package bundles

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/google/uuid"
)

type BundleArchiveWriter interface {
	Open() error
	WriteChartArchive(data []byte) error
	WriteImageArchive(imageTag string, data []byte) error
	Save() error
}

type BundleArchiveFileWriter struct {
	Path string

	tmpArchivePath   string
	tmpArchiveWriter *tar.Writer
	tmpArchiveCloser func() error
}

func NewBundleArchiveFileWriter(path string) *BundleArchiveFileWriter {
	return &BundleArchiveFileWriter{Path: path}
}

func (writer *BundleArchiveFileWriter) Open() error {
	p := fmt.Sprintf("%s.%s.tmp", writer.Path, uuid.New().String())

	f, err := os.Create(p)
	if err != nil {
		return fmt.Errorf("unable to open tmp archive file %q: %w", p, err)
	}

	zipper := gzip.NewWriter(f)
	zipper.Header.Comment = "bundle-archive"
	twriter := tar.NewWriter(zipper)

	writer.tmpArchivePath = p
	writer.tmpArchiveWriter = twriter
	writer.tmpArchiveCloser = func() error {
		if err := twriter.Close(); err != nil {
			return fmt.Errorf("unable to close tar writer for %q: %w", writer.tmpArchivePath, err)
		}
		if err := zipper.Close(); err != nil {
			return fmt.Errorf("unable to close zipper for %q: %w", writer.tmpArchivePath, err)
		}
		if err := f.Close(); err != nil {
			return fmt.Errorf("unable to close %q: %w", writer.tmpArchivePath, err)
		}
		return nil
	}

	now := time.Now()
	header := &tar.Header{
		Name:       "images",
		Typeflag:   tar.TypeDir,
		Mode:       0o777,
		ModTime:    now,
		AccessTime: now,
		ChangeTime: now,
	}
	if err := writer.tmpArchiveWriter.WriteHeader(header); err != nil {
		return fmt.Errorf("unable to write images dir header: %w", err)
	}

	return nil
}

func (writer *BundleArchiveFileWriter) Save() error {
	if writer.tmpArchiveWriter == nil {
		panic(fmt.Sprintf("bundle archive %q is not opened", writer.Path))
	}

	if err := writer.tmpArchiveCloser(); err != nil {
		return fmt.Errorf("unable to close tmp archive %q: %w", writer.tmpArchivePath, err)
	}

	if err := os.RemoveAll(writer.Path); err != nil {
		return fmt.Errorf("unable to cleanup destination archive path %q: %w", writer.Path, err)
	}

	if err := os.Rename(writer.tmpArchivePath, writer.Path); err != nil {
		return fmt.Errorf("unable to rename tmp bundle archive %q to %q: %w", writer.tmpArchivePath, writer.Path, err)
	}

	return nil
}

func (writer *BundleArchiveFileWriter) WriteChartArchive(data []byte) error {
	now := time.Now()
	header := &tar.Header{
		Name:       chartArchiveFileName,
		Typeflag:   tar.TypeReg,
		Mode:       0o777,
		Size:       int64(len(data)),
		ModTime:    now,
		AccessTime: now,
		ChangeTime: now,
	}

	if err := writer.tmpArchiveWriter.WriteHeader(header); err != nil {
		return fmt.Errorf("unable to write %q header: %w", chartArchiveFileName, err)
	}

	if _, err := writer.tmpArchiveWriter.Write(data); err != nil {
		return fmt.Errorf("unable to write %q data: %w", chartArchiveFileName, err)
	}

	return nil
}

func (writer *BundleArchiveFileWriter) WriteImageArchive(imageTag string, data []byte) error {
	now := time.Now()
	buf := bytes.NewBuffer(nil)
	zipper := gzip.NewWriter(buf)

	if _, err := io.Copy(zipper, bytes.NewReader(data)); err != nil {
		return fmt.Errorf("unable to gzip image archive data: %w", err)
	}

	if err := zipper.Close(); err != nil {
		return fmt.Errorf("unable to close gzip image archive: %w", err)
	}

	header := &tar.Header{
		Name:       fmt.Sprintf("images/%s.tar.gz", imageTag),
		Typeflag:   tar.TypeReg,
		Mode:       0o777,
		Size:       int64(len(buf.Bytes())),
		ModTime:    now,
		AccessTime: now,
		ChangeTime: now,
	}

	if err := writer.tmpArchiveWriter.WriteHeader(header); err != nil {
		return fmt.Errorf("unable to write image %q header: %w", imageTag, err)
	}

	if _, err := writer.tmpArchiveWriter.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("unable to write chart.tar.gz data: %w", err)
	}

	return nil
}
