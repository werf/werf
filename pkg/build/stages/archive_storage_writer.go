package stages

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

type ArchiveStorageWriter interface {
	Open() error
	Save() error
	WriteStageArchive(stageTag string, data []byte) error
}

type ArchiveStorageFileWriteCloser struct {
	Path string

	tmpArchivePath   string
	tmpArchiveWriter *tar.Writer
	tmpArchiveCloser func() error
}

func NewArchiveStorageFileWriter(path string) *ArchiveStorageFileWriteCloser {
	return &ArchiveStorageFileWriteCloser{
		Path: path,
	}
}

func (writer *ArchiveStorageFileWriteCloser) Open() error {
	p := fmt.Sprintf("%s.%s.tmp", writer.Path, uuid.New().String())

	f, err := os.Create(p)
	if err != nil {
		return fmt.Errorf("unable to open tmp archive file %q: %w", p, err)
	}

	zipper := gzip.NewWriter(f)
	zipper.Header.Comment = "stage-archive"
	tWriter := tar.NewWriter(zipper)

	writer.tmpArchivePath = p
	writer.tmpArchiveWriter = tWriter
	writer.tmpArchiveCloser = func() error {
		if err := tWriter.Close(); err != nil {
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
		Name:       "stages",
		Typeflag:   tar.TypeDir,
		Mode:       0o777,
		ModTime:    now,
		AccessTime: now,
		ChangeTime: now,
	}

	if err := writer.tmpArchiveWriter.WriteHeader(header); err != nil {
		return fmt.Errorf("unable to write stages dir header: %w", err)
	}

	return nil
}

func (writer *ArchiveStorageFileWriteCloser) Save() error {
	if writer.tmpArchiveWriter == nil {
		return fmt.Errorf("stage archive %q is not opened", writer.Path)
	}

	if err := writer.tmpArchiveCloser(); err != nil {
		return fmt.Errorf("unable to close tmp archive %q: %w", writer.tmpArchivePath, err)
	}

	if err := os.RemoveAll(writer.Path); err != nil {
		return fmt.Errorf("unable to cleanup destination archive path %q: %w", writer.Path, err)
	}

	if err := os.Rename(writer.tmpArchivePath, writer.Path); err != nil {
		return fmt.Errorf("unable to rename tmp stage archive %q to %q: %w", writer.tmpArchivePath, writer.Path, err)
	}

	return nil
}

func (writer *ArchiveStorageFileWriteCloser) WriteStageArchive(tag string, data []byte) error {
	now := time.Now()
	buf := bytes.NewBuffer(nil)
	zipper := gzip.NewWriter(buf)

	if _, err := io.Copy(zipper, bytes.NewReader(data)); err != nil {
		return fmt.Errorf("unable to gzip image archive data: %w", err)
	}

	if err := zipper.Close(); err != nil {
		return fmt.Errorf("unable to close gzip image archive: %w", err)
	}

	compressedData := buf.Bytes()

	header := &tar.Header{
		Name:       fmt.Sprintf("stages/%s.tar.gz", tag),
		Typeflag:   tar.TypeReg,
		Mode:       0o777,
		Size:       int64(len(buf.Bytes())),
		ModTime:    now,
		AccessTime: now,
		ChangeTime: now,
	}

	if err := writer.tmpArchiveWriter.WriteHeader(header); err != nil {
		return fmt.Errorf("unable to write stage %q header: %w", tag, err)
	}

	if _, err := writer.tmpArchiveWriter.Write(compressedData); err != nil {
		return fmt.Errorf("unable to write stage.tar.gz data: %w", err)
	}

	return nil
}
