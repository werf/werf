package stages

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/google/uuid"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/tmp_manager"
)

type ArchiveStorageWriter interface {
	WriteStageArchive(stageTag string, pull func(writer io.Writer) error) error
	WithTask(ctx context.Context, task func(ArchiveStorageWriter) error) error
}

type ArchiveStorageFileWriter struct {
	Path string

	tmpArchivePath   string
	tmpArchiveWriter *tar.Writer
	tmpArchiveCloser func() error
}

func NewArchiveStorageFileWriter(path string) *ArchiveStorageFileWriter {
	return &ArchiveStorageFileWriter{
		Path: path,
	}
}

func (writer *ArchiveStorageFileWriter) open() error {
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

func (writer *ArchiveStorageFileWriter) save() error {
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

func (writer *ArchiveStorageFileWriter) writeStageArchive(tag string, reader io.Reader) error {
	tmpFile, err := tmp_manager.TempFile("stage-archive-*.tar.gz")
	if err != nil {
		return fmt.Errorf("unable to create temp file for stage archive: %w", err)
	}

	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	zipper := gzip.NewWriter(tmpFile)
	if _, err := io.Copy(zipper, reader); err != nil {
		tmpFile.Close()
		return fmt.Errorf("unable to gzip stage archive data: %w", err)
	}
	if err := zipper.Close(); err != nil {
		tmpFile.Close()
		return fmt.Errorf("unable to close gzip writer: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("unable to close temp file: %w", err)
	}

	stat, err := os.Stat(tmpPath)
	if err != nil {
		return fmt.Errorf("unable to stat temp file: %w", err)
	}

	now := time.Now()
	header := &tar.Header{
		Name:       fmt.Sprintf(stagePathTemplate, tag),
		Typeflag:   tar.TypeReg,
		Mode:       0o777,
		Size:       stat.Size(),
		ModTime:    now,
		AccessTime: now,
		ChangeTime: now,
	}

	if err := writer.tmpArchiveWriter.WriteHeader(header); err != nil {
		return fmt.Errorf("unable to write stage %q header: %w", tag, err)
	}

	compressedFile, err := os.Open(tmpPath)
	if err != nil {
		return fmt.Errorf("unable to open temp file for reading: %w", err)
	}
	defer compressedFile.Close()

	if _, err := io.Copy(writer.tmpArchiveWriter, compressedFile); err != nil {
		return fmt.Errorf("unable to write stage data: %w", err)
	}

	return nil
}

func (writer *ArchiveStorageFileWriter) WriteStageArchive(tag string, pull func(w io.Writer) error) error {
	tmpFile, err := tmp_manager.TempFile("stage-pull-*.tar")
	if err != nil {
		return fmt.Errorf("unable to create temp file for pulling stage: %w", err)
	}

	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if err := pull(tmpFile); err != nil {
		tmpFile.Close()
		return fmt.Errorf("unable to pull stage data: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("unable to close temp file: %w", err)
	}

	f, err := os.Open(tmpPath)
	if err != nil {
		return fmt.Errorf("unable to open temp file for reading: %w", err)
	}
	defer f.Close()

	return writer.writeStageArchive(tag, f)
}

func (writer *ArchiveStorageFileWriter) Close() error {
	if writer.tmpArchiveCloser != nil {
		return writer.tmpArchiveCloser()
	}

	return nil
}

func (writer *ArchiveStorageFileWriter) WithTask(ctx context.Context, task func(ArchiveStorageWriter) error) error {
	if err := writer.open(); err != nil {
		return fmt.Errorf("unable to open target stages archive: %w", err)
	}

	var err error
	defer func() {
		if err != nil {
			if closeErr := writer.Close(); closeErr != nil {
				logboek.Context(ctx).Warn().LogF("Warning: error closing archive after task failure: %v\n", closeErr)
			}
			if writer.tmpArchivePath != "" {
				os.Remove(writer.tmpArchivePath)
			}
		}
	}()

	err = task(writer)
	if err != nil {
		return err
	}

	if err := writer.save(); err != nil {
		return fmt.Errorf("error saving destination bundle archive: %w", err)
	}

	return nil
}
